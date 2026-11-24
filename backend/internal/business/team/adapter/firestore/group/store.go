package firestore

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"time"

	db "cloud.google.com/go/firestore"

	teamdb "github.com/963sup/line-mini-app-bot/internal/business/team/adapter/firestore"
	groupdto "github.com/963sup/line-mini-app-bot/internal/business/team/application/dto"
	groupport "github.com/963sup/line-mini-app-bot/internal/business/team/application/port"
	groupentity "github.com/963sup/line-mini-app-bot/internal/business/team/domain/entity"
	groupevent "github.com/963sup/line-mini-app-bot/internal/business/team/domain/event"
	"github.com/963sup/line-mini-app-bot/internal/platform/firestore/key"
	"github.com/963sup/line-mini-app-bot/internal/platform/line/lineid"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type Store struct {
	Client    *db.Client
	ChannelID string
}

type document struct {
	ID        string    `firestore:"id"`
	GroupID   string    `firestore:"groupId"`
	Name      string    `firestore:"name"`
	Status    string    `firestore:"status"`
	Version   int64     `firestore:"version"`
	CreatedAt time.Time `firestore:"createdAt"`
	UpdatedAt time.Time `firestore:"updatedAt"`
	EventAt   time.Time `firestore:"eventAt"`
	EventID   string    `firestore:"eventId"`
}

func (d document) model() groupentity.Group {
	return groupentity.Group{ID: d.ID, GroupID: d.GroupID, Name: d.Name, Status: d.Status, Version: d.Version, CreatedAt: d.CreatedAt, UpdatedAt: d.UpdatedAt}
}

func (s *Store) channel() *db.DocumentRef { return s.Client.Collection("channels").Doc(s.ChannelID) }

// The LINE group ID is the document key. Event
// ordering prevents old deliveries from reactivating a departed group.
func (s *Store) Observe(ctx context.Context, o groupevent.Observation) error {
	if !lineid.ValidGroup(o.GroupID) || o.At.IsZero() || o.EventID == "" || (o.Status != "ACTIVE" && o.Status != "INACTIVE") {
		return groupport.ErrUnavailable
	}
	id, now := o.GroupID, time.Now().UTC()
	return s.Client.RunTransaction(ctx, func(_ context.Context, tx *db.Transaction) error {
		ref := s.channel().Collection("groups").Doc(id)
		d := document{ID: id, GroupID: o.GroupID, Status: o.Status, Name: o.Name, Version: 1, CreatedAt: now, UpdatedAt: now, EventAt: o.At, EventID: o.EventID}
		writeGroup := true
		existing, readErr := tx.Get(ref)
		if readErr != nil && status.Code(readErr) != codes.NotFound {
			return readErr
		}
		if readErr == nil {
			var err error
			d, err = decodeDocument(existing)
			if err != nil {
				return err
			}
			// Equal timestamps prefer INACTIVE; an exact replay never changes version.
			outdated := o.At.Before(d.EventAt)
			replayOrInactive := o.At.Equal(d.EventAt) && (o.EventID == d.EventID || d.Status == "INACTIVE")
			writeGroup = !outdated && !replayOrInactive
			if writeGroup {
				d.EventAt, d.EventID = o.At, o.EventID
				if d.Status != o.Status || (o.Name != "" && d.Name != o.Name) {
					d.Status, d.UpdatedAt = o.Status, now
					d.Version++
					if o.Name != "" {
						d.Name = o.Name
					}
				}
			}
		}
		if writeGroup {
			if err := tx.Set(ref, d); err != nil {
				return err
			}
		}
		// Apply the accepted observation even on replay to backfill existing teams.
		if err := teamdb.ApplyGroupObservation(tx, s.channel(), id, d.Name, d.Status); err != nil {
			return err
		}
		seen := map[string]bool{}
		for _, user := range o.Users {
			if user == "" || seen[user] {
				continue
			}
			seen[user] = true
			candidate := s.channel().Collection("groupUsers").Doc(firestorekey.Digest([]byte(user))).Collection("groups").Doc(id)
			if err := tx.Set(candidate, map[string]any{"groupId": id}); err != nil {
				return err
			}
		}
		return nil
	})
}

func decode(snap *db.DocumentSnapshot, err error) (groupentity.Group, error) {
	if status.Code(err) == codes.NotFound {
		return groupentity.Group{}, groupentity.ErrNotFound
	}
	if err != nil {
		return groupentity.Group{}, err
	}
	d, err := decodeDocument(snap)
	return d.model(), err
}

func decodeDocument(snap *db.DocumentSnapshot) (document, error) {
	var d document
	if err := snap.DataTo(&d); err != nil {
		return document{}, err
	}
	if !lineid.ValidGroup(d.ID) || d.ID != d.GroupID || d.ID != snap.Ref.ID || d.Version < 1 ||
		(d.Status != "ACTIVE" && d.Status != "INACTIVE") {
		return document{}, groupport.ErrUnavailable
	}
	return d, nil
}

func (s *Store) GetByID(ctx context.Context, id string) (groupentity.Group, error) {
	if !lineid.ValidGroup(id) {
		return groupentity.Group{}, groupentity.ErrNotFound
	}
	return decode(s.channel().Collection("groups").Doc(id).Get(ctx))
}

type cursor struct{ Scope, ID string }

func (s *Store) Candidates(ctx context.Context, user, value string) (groupdto.Page, error) {
	page := groupdto.Page{Items: []groupentity.Group{}}
	scope := firestorekey.Digest([]byte(s.ChannelID + ":" + user))
	c := cursor{}
	if value != "" {
		b, err := base64.RawURLEncoding.DecodeString(value)
		if len(value) > 1024 || err != nil || json.Unmarshal(b, &c) != nil || c.Scope != scope || !lineid.ValidGroup(c.ID) {
			return page, groupport.ErrCursor
		}
	}
	q := s.channel().Collection("groupUsers").Doc(firestorekey.Digest([]byte(user))).Collection("groups").OrderBy(db.DocumentID, db.Asc)
	if c.ID != "" {
		q = q.StartAfter(c.ID)
	}
	docs, err := q.Limit(11).Documents(ctx).GetAll()
	if err != nil {
		return page, err
	}
	if len(docs) > 10 {
		b, err := json.Marshal(cursor{Scope: scope, ID: docs[9].Ref.ID})
		if err != nil {
			return page, err
		}
		page.NextCursor = base64.RawURLEncoding.EncodeToString(b)
		docs = docs[:10]
	}
	for _, doc := range docs {
		r, err := s.GetByID(ctx, doc.Ref.ID)
		if err != nil {
			return page, err
		}
		page.Items = append(page.Items, r)
	}
	return page, nil
}
