package firestore

import (
	"context"
	"strconv"
	"strings"
	"time"

	db "cloud.google.com/go/firestore"
	"github.com/963sup/line-mini-app-bot/internal/business/object/application/dto"
	"github.com/963sup/line-mini-app-bot/internal/business/object/domain/entity"
	repository "github.com/963sup/line-mini-app-bot/internal/business/repository/domain/entity"
	key "github.com/963sup/line-mini-app-bot/internal/platform/firestore/key"
	"google.golang.org/api/iterator"
)

func (s *Store) ListQuantityChanges(ctx context.Context, r repository.Repository, actor, token string) (page dto.QuantityChangePage, err error) {
	defer func() { err = translate(err) }()
	if !s.ready() {
		return page, entity.ErrUnavailable
	}
	repo := s.repositoryRef(r.ID)
	q := s.Client.CollectionGroup("changes").Where("after.repositoryId", "==", r.ID).OrderBy("at", db.Desc).OrderBy(db.DocumentID, db.Desc)
	scope := []string{"quantity-history", s.ChannelID, r.ID, actor}
	if token != "" {
		last, ok := key.DecodeCursor(token, scope)
		id, v, found := strings.Cut(last, "/")
		version, e := strconv.ParseInt(v, 10, 64)
		if !ok || !found || !repository.ValidID(id) || e != nil || version < 2 || strconv.FormatInt(version, 10) != v {
			return page, entity.ErrInput
		}
		snapshot, e := repo.Collection("objects").Doc(id).Collection("changes").Doc(v).Get(ctx)
		if e != nil {
			return page, e
		}
		q = q.StartAfter(snapshot)
	}
	const limit = 25
	iter := q.Documents(ctx)
	defer iter.Stop()
	var last string
	scanned := 0
	for {
		snapshot, e := iter.Next()
		if e == iterator.Done {
			break
		}
		if e != nil {
			return page, e
		}
		object := snapshot.Ref.Parent.Parent
		if object == nil || object.Parent.Path != repo.Collection("objects").Path {
			continue
		}
		// Bound reads even when a repository has many name-only changes.
		if scanned == 100 || len(page.Items) == limit {
			page.NextCursor = key.EncodeCursor(scope, last)
			break
		}
		scanned++
		last = object.ID + "/" + snapshot.Ref.ID
		var c struct {
			ActorName string    `firestore:"actorName"`
			ActorID   string    `firestore:"actorId"`
			Version   int64     `firestore:"version"`
			At        time.Time `firestore:"at"`
			Before    document  `firestore:"before"`
			After     document  `firestore:"after"`
		}
		if err = snapshot.DataTo(&c); err != nil {
			return page, err
		}
		ref := repo.Collection("objects").Doc(c.After.ID).Collection("changes").Doc(strconv.FormatInt(c.Version, 10))
		if snapshot.Ref.Path != ref.Path || c.Before.RepositoryID != r.ID || c.After.RepositoryID != r.ID {
			return page, entity.ErrUnavailable
		}
		if c.Before.Quantity == c.After.Quantity {
			continue
		}
		label := strings.TrimSpace(c.ActorName)
		if label == "" {
			if c.ActorID == actor {
				label = "本人"
			} else {
				label = "成員 " + key.JSONDigest([]string{s.ChannelID, r.ID, c.ActorID})[:12]
			}
		}
		page.Items = append(page.Items, dto.QuantityChange{ObjectID: c.After.ID, Name: c.After.Name, Version: c.Version, At: c.At, Actor: label, BeforeQuantity: c.Before.Quantity, BeforeUnit: c.Before.Unit, AfterQuantity: c.After.Quantity, AfterUnit: c.After.Unit})
	}
	return page, nil
}
