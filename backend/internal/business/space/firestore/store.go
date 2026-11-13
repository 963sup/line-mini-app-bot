package firestore

import (
	"context"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"time"

	db "cloud.google.com/go/firestore"
	"github.com/963sup/line-mini-app-bot/internal/business/space"
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

func (d document) model() space.Space {
	return space.Space{ID: d.ID, GroupID: d.GroupID, Name: d.Name, Status: d.Status, Version: d.Version, CreatedAt: d.CreatedAt, UpdatedAt: d.UpdatedAt}
}

func (s *Store) channel() *db.DocumentRef { return s.Client.Collection("channels").Doc(s.ChannelID) }

func key(value string) string { sum := sha256.Sum256([]byte(value)); return hex.EncodeToString(sum[:]) }

// Observe commits the independent ID and unique group binding together. Event
// ordering prevents old deliveries from reactivating a departed group.
func (s *Store) Observe(ctx context.Context, o space.Observation) error {
	if !space.ValidGroup(o.GroupID) || o.At.IsZero() || o.EventID == "" || (o.Status != "ACTIVE" && o.Status != "INACTIVE") {
		return space.ErrUnavailable
	}
	newID, now := o.GroupID, time.Now().UTC()
	binding := s.channel().Collection("spaceGroupBindings").Doc(o.GroupID)
	return s.Client.RunTransaction(ctx, func(_ context.Context, tx *db.Transaction) error {
		id := newID
		bound, err := tx.Get(binding)
		if err == nil {
			var b struct {
				ID string `firestore:"spaceId"`
			}
			if err := bound.DataTo(&b); err != nil {
				return err
			}
			if !space.ValidID(b.ID) || b.ID != o.GroupID {
				return space.ErrUnavailable
			}
			id = b.ID
		} else if status.Code(err) != codes.NotFound {
			return err
		}
		ref := s.channel().Collection("spaces").Doc(id)
		d := document{ID: id, GroupID: o.GroupID, Status: o.Status, Name: o.Name, Version: 1, CreatedAt: now, UpdatedAt: now, EventAt: o.At, EventID: o.EventID}
		writeRepository := true
		existing, readErr := tx.Get(ref)
		if readErr != nil && status.Code(readErr) != codes.NotFound {
			return readErr
		}
		if readErr == nil {
			snap := existing
			if err := snap.DataTo(&d); err != nil {
				return err
			}
			if d.ID != id || d.GroupID != o.GroupID || d.Version < 1 || (d.Status != "ACTIVE" && d.Status != "INACTIVE") {
				return space.ErrUnavailable
			}
			// Equal timestamps prefer INACTIVE; an exact replay never changes version.
			outdated := o.At.Before(d.EventAt)
			replayOrInactive := o.At.Equal(d.EventAt) && (o.EventID == d.EventID || d.Status == "INACTIVE")
			writeRepository = !outdated && !replayOrInactive
			if writeRepository {
				d.EventAt, d.EventID = o.At, o.EventID
				if d.Status != o.Status || (o.Name != "" && d.Name != o.Name) {
					d.Status, d.UpdatedAt = o.Status, now
					d.Version++
					if o.Name != "" {
						d.Name = o.Name
					}
				}
			}
		} else if bound != nil && bound.Exists() {
			return space.ErrUnavailable
		}
		if bound == nil || !bound.Exists() {
			if err := tx.Create(binding, map[string]any{"spaceId": id}); err != nil {
				return err
			}
		}
		if writeRepository {
			if err := tx.Set(ref, d); err != nil {
				return err
			}
		}
		seen := map[string]bool{}
		for _, user := range o.Users {
			if user == "" || seen[user] {
				continue
			}
			seen[user] = true
			candidate := s.channel().Collection("spaceUsers").Doc(key(user)).Collection("spaces").Doc(id)
			if err := tx.Set(candidate, map[string]any{"spaceId": id}); err != nil {
				return err
			}
		}
		return nil
	})
}

func decode(snap *db.DocumentSnapshot, err error) (space.Space, error) {
	if status.Code(err) == codes.NotFound {
		return space.Space{}, space.ErrNotFound
	}
	if err != nil {
		return space.Space{}, err
	}
	var d document
	if err := snap.DataTo(&d); err != nil {
		return space.Space{}, err
	}
	if !space.ValidID(d.ID) || d.ID != d.GroupID || d.ID != snap.Ref.ID || !space.ValidGroup(d.GroupID) || d.Version < 1 {
		return space.Space{}, space.ErrUnavailable
	}
	return d.model(), nil
}

func (s *Store) GetByID(ctx context.Context, id string) (space.Space, error) {
	if !space.ValidID(id) {
		return space.Space{}, space.ErrNotFound
	}
	return decode(s.channel().Collection("spaces").Doc(id).Get(ctx))
}

func (s *Store) GetByGroupID(ctx context.Context, group string) (space.Space, error) {
	if !space.ValidGroup(group) {
		return space.Space{}, space.ErrNotFound
	}
	snap, err := s.channel().Collection("spaceGroupBindings").Doc(group).Get(ctx)
	if status.Code(err) == codes.NotFound {
		return space.Space{}, space.ErrNotFound
	}
	if err != nil {
		return space.Space{}, err
	}
	var b struct {
		ID string `firestore:"spaceId"`
	}
	if err := snap.DataTo(&b); err != nil {
		return space.Space{}, err
	}
	r, err := s.GetByID(ctx, b.ID)
	if err == nil && r.GroupID != group {
		return space.Space{}, space.ErrUnavailable
	}
	return r, err
}

type cursor struct{ Scope, ID string }

func (s *Store) Candidates(ctx context.Context, user, value string) (space.Page, error) {
	page := space.Page{Items: []space.Space{}}
	scope := key(s.ChannelID + ":" + user)
	c := cursor{}
	if value != "" {
		b, err := base64.RawURLEncoding.DecodeString(value)
		if len(value) > 1024 || err != nil || json.Unmarshal(b, &c) != nil || c.Scope != scope || !space.ValidID(c.ID) {
			return page, space.ErrCursor
		}
	}
	q := s.channel().Collection("spaceUsers").Doc(key(user)).Collection("spaces").OrderBy(db.DocumentID, db.Asc)
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
