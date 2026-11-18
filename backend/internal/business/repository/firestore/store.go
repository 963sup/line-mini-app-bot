package firestore

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"time"

	db "cloud.google.com/go/firestore"
	"github.com/963sup/line-mini-app-bot/internal/business/repository"
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

func (d document) model() repository.Repository {
	return repository.Repository{ID: d.ID, GroupID: d.GroupID, Name: d.Name, Status: d.Status, Version: d.Version, CreatedAt: d.CreatedAt, UpdatedAt: d.UpdatedAt}
}

func (s *Store) channel() *db.DocumentRef { return s.Client.Collection("channels").Doc(s.ChannelID) }

func key(value string) string { sum := sha256.Sum256([]byte(value)); return hex.EncodeToString(sum[:]) }

// Observe commits the independent ID and unique group binding together. Event
// ordering prevents old deliveries from reactivating a departed group.
func (s *Store) Observe(ctx context.Context, o repository.Observation) error {
	if !repository.ValidGroup(o.GroupID) || o.At.IsZero() || o.EventID == "" || (o.Status != "ACTIVE" && o.Status != "INACTIVE") {
		return repository.ErrUnavailable
	}
	var random [16]byte
	if _, err := rand.Read(random[:]); err != nil {
		return err
	}
	newID, now := hex.EncodeToString(random[:]), time.Now().UTC()
	binding := s.channel().Collection("repositoryGroupBindings").Doc(o.GroupID)
	return s.Client.RunTransaction(ctx, func(_ context.Context, tx *db.Transaction) error {
		id := newID
		bound, err := tx.Get(binding)
		if err == nil {
			var b struct {
				ID string `firestore:"repositoryId"`
			}
			if err := bound.DataTo(&b); err != nil {
				return err
			}
			if !repository.ValidID(b.ID) {
				return repository.ErrUnavailable
			}
			id = b.ID
		} else if status.Code(err) != codes.NotFound {
			return err
		}
		ref := s.channel().Collection("repositories").Doc(id)
		d := document{ID: id, GroupID: o.GroupID, Status: o.Status, Name: o.Name, Version: 1, CreatedAt: now, UpdatedAt: now, EventAt: o.At, EventID: o.EventID}
		writeRepository := true
		if bound != nil && bound.Exists() {
			snap, err := tx.Get(ref)
			if err != nil {
				return err
			}
			if err := snap.DataTo(&d); err != nil {
				return err
			}
			if d.ID != id || d.GroupID != o.GroupID {
				return repository.ErrUnavailable
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
		} else {
			if err := tx.Create(binding, map[string]any{"repositoryId": id}); err != nil {
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
			candidate := s.channel().Collection("repositoryUsers").Doc(key(user)).Collection("repositories").Doc(id)
			if err := tx.Set(candidate, map[string]any{"repositoryId": id}); err != nil {
				return err
			}
		}
		return nil
	})
}

func decode(snap *db.DocumentSnapshot, err error) (repository.Repository, error) {
	if status.Code(err) == codes.NotFound {
		return repository.Repository{}, repository.ErrNotFound
	}
	if err != nil {
		return repository.Repository{}, err
	}
	var d document
	if err := snap.DataTo(&d); err != nil {
		return repository.Repository{}, err
	}
	if !repository.ValidID(d.ID) || d.ID != snap.Ref.ID || !repository.ValidGroup(d.GroupID) || d.Version < 1 {
		return repository.Repository{}, repository.ErrUnavailable
	}
	return d.model(), nil
}

func (s *Store) GetByID(ctx context.Context, id string) (repository.Repository, error) {
	if !repository.ValidID(id) {
		return repository.Repository{}, repository.ErrNotFound
	}
	return decode(s.channel().Collection("repositories").Doc(id).Get(ctx))
}

func (s *Store) GetByGroupID(ctx context.Context, group string) (repository.Repository, error) {
	if !repository.ValidGroup(group) {
		return repository.Repository{}, repository.ErrNotFound
	}
	snap, err := s.channel().Collection("repositoryGroupBindings").Doc(group).Get(ctx)
	if status.Code(err) == codes.NotFound {
		return repository.Repository{}, repository.ErrNotFound
	}
	if err != nil {
		return repository.Repository{}, err
	}
	var b struct {
		ID string `firestore:"repositoryId"`
	}
	if err := snap.DataTo(&b); err != nil {
		return repository.Repository{}, err
	}
	r, err := s.GetByID(ctx, b.ID)
	if err == nil && r.GroupID != group {
		return repository.Repository{}, repository.ErrUnavailable
	}
	return r, err
}

type cursor struct{ Scope, ID string }

func (s *Store) Candidates(ctx context.Context, user, value string) (repository.Page, error) {
	page := repository.Page{Items: []repository.Repository{}}
	scope := key(s.ChannelID + ":" + user)
	c := cursor{}
	if value != "" {
		b, err := base64.RawURLEncoding.DecodeString(value)
		if len(value) > 1024 || err != nil || json.Unmarshal(b, &c) != nil || c.Scope != scope || !repository.ValidID(c.ID) {
			return page, repository.ErrCursor
		}
	}
	q := s.channel().Collection("repositoryUsers").Doc(key(user)).Collection("repositories").OrderBy(db.DocumentID, db.Asc)
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
