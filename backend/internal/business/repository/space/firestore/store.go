package firestore

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"errors"
	"time"

	db "cloud.google.com/go/firestore"
	"github.com/963sup/line-mini-app-bot/internal/business/repository"
	feature "github.com/963sup/line-mini-app-bot/internal/business/repository/space"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type Store struct {
	Client    *db.Client
	ChannelID string
}

var _ feature.Store = (*Store)(nil)

type document struct {
	ID            string    `firestore:"id"`
	RepositoryID  string    `firestore:"repositoryId"`
	Name          string    `firestore:"name"`
	CreatorID     string    `firestore:"creatorId"`
	SchemaVersion int64     `firestore:"schemaVersion"`
	Version       int64     `firestore:"version"`
	CreatedAt     time.Time `firestore:"createdAt"`
	UpdatedAt     time.Time `firestore:"updatedAt"`
}

func encode(s feature.Space) document {
	return document{ID: s.ID, RepositoryID: s.RepositoryID, Name: s.Name, CreatorID: s.CreatorID, SchemaVersion: 1, Version: s.Version, CreatedAt: s.CreatedAt, UpdatedAt: s.UpdatedAt}
}

func decode(snap *db.DocumentSnapshot, err error, rid string) (feature.Space, error) {
	if status.Code(err) == codes.NotFound {
		return feature.Space{}, feature.ErrNotFound
	}
	if err != nil {
		return feature.Space{}, err
	}
	var d document
	if e := snap.DataTo(&d); e != nil {
		return feature.Space{}, e
	}
	n, e := feature.NormalizeName(d.Name)
	if e != nil || n != d.Name || d.ID != snap.Ref.ID || !repository.ValidID(d.ID) || d.RepositoryID != rid || d.SchemaVersion != 1 || d.CreatorID == "" || d.Version != 1 || d.CreatedAt.IsZero() || d.UpdatedAt.Before(d.CreatedAt) {
		return feature.Space{}, feature.ErrUnavailable
	}
	return feature.Space{ID: d.ID, RepositoryID: d.RepositoryID, Name: d.Name, CreatorID: d.CreatorID, Version: d.Version, CreatedAt: d.CreatedAt, UpdatedAt: d.UpdatedAt}, nil
}
func (s *Store) ready() bool { return s != nil && s.Client != nil && s.ChannelID != "" }
func (s *Store) parent(rid string) *db.DocumentRef {
	return s.Client.Collection("channels").Doc(s.ChannelID).Collection("repositories").Doc(rid)
}

func hash(v any) string {
	b, _ := json.Marshal(v)
	h := sha256.Sum256(b)
	return hex.EncodeToString(h[:])
}

func translate(e error) error {
	if e == nil {
		return nil
	}
	for _, v := range []error{feature.ErrInput, feature.ErrNotFound, feature.ErrCreationConflict, feature.ErrUnavailable, context.Canceled, context.DeadlineExceeded} {
		if errors.Is(e, v) {
			return e
		}
	}
	switch status.Code(e) {
	case codes.Canceled:
		return context.Canceled
	case codes.DeadlineExceeded:
		return context.DeadlineExceeded
	}
	return feature.ErrUnavailable
}

func (s *Store) Get(ctx context.Context, r repository.Repository, id string) (result feature.Space, err error) {
	defer func() { err = translate(err) }()
	if !s.ready() {
		return result, feature.ErrUnavailable
	}
	snap, e := s.parent(r.ID).Collection("spaces").Doc(id).Get(ctx)
	return decode(snap, e, r.ID)
}

type cursor struct{ Scope, ID string }

func (s *Store) List(ctx context.Context, r repository.Repository, actor, token string) (result feature.Page, err error) {
	defer func() { err = translate(err) }()
	result.Items = []feature.Space{}
	if !s.ready() {
		return result, feature.ErrUnavailable
	}
	scope := hash([]string{s.ChannelID, r.ID, actor})
	c := cursor{}
	if token != "" {
		b, e := base64.RawURLEncoding.DecodeString(token)
		if e != nil || json.Unmarshal(b, &c) != nil || c.Scope != scope || !repository.ValidID(c.ID) {
			return result, feature.ErrInput
		}
	}
	q := s.parent(r.ID).Collection("spaces").OrderBy(db.DocumentID, db.Asc)
	if c.ID != "" {
		q = q.StartAfter(c.ID)
	}
	docs, e := q.Limit(21).Documents(ctx).GetAll()
	if e != nil {
		return result, e
	}
	more := len(docs) > 20
	if more {
		docs = docs[:20]
	}
	for _, snap := range docs {
		i, e := decode(snap, nil, r.ID)
		if e != nil {
			return feature.Page{}, e
		}
		result.Items = append(result.Items, i)
		c = cursor{scope, i.ID}
	}
	if more {
		b, _ := json.Marshal(c)
		result.NextCursor = base64.RawURLEncoding.EncodeToString(b)
	}
	return result, nil
}

func container(tx *db.Transaction, parent *db.DocumentRef, r repository.Repository) error {
	snap, e := tx.Get(parent)
	if status.Code(e) == codes.NotFound {
		return feature.ErrNotFound
	}
	if e != nil {
		return e
	}
	var d struct {
		ID      string `firestore:"id"`
		OwnerID string `firestore:"ownerId"`
		Status  string `firestore:"status"`
	}
	if e = snap.DataTo(&d); e != nil {
		return e
	}
	if d.ID != r.ID || d.OwnerID != r.OwnerID || d.Status != "ACTIVE" {
		return feature.ErrNotFound
	}
	return nil
}

type receipt struct {
	SchemaVersion int64  `firestore:"schemaVersion"`
	ID            string `firestore:"id"`
	PayloadHash   string `firestore:"payloadHash"`
}

func (s *Store) Create(ctx context.Context, r repository.Repository, actor, key string, name string) (result feature.Space, err error) {
	defer func() { err = translate(err) }()
	if !s.ready() {
		return result, feature.ErrUnavailable
	}
	var random [16]byte
	if _, e := rand.Read(random[:]); e != nil {
		return result, e
	}
	id := hex.EncodeToString(random[:])
	now := time.Now().UTC().Truncate(time.Microsecond)
	parent := s.parent(r.ID)
	ref := parent.Collection("spaces").Doc(id)
	// Space receipts do not allocate repository resource numbers.
	rec := parent.Collection("spaceCreationReceipts").Doc(hash([]string{actor, key}))
	payload := hash(struct{ Name string }{name})
	err = s.Client.RunTransaction(ctx, func(_ context.Context, tx *db.Transaction) error {
		result = feature.Space{}
		if e := container(tx, parent, r); e != nil {
			return e
		}
		prior, e := tx.Get(rec)
		if e == nil {
			var old receipt
			if e = prior.DataTo(&old); e != nil {
				return e
			}
			if old.SchemaVersion != 1 || !repository.ValidID(old.ID) || len(old.PayloadHash) != 64 {
				return feature.ErrUnavailable
			}
			if old.PayloadHash != payload {
				return feature.ErrCreationConflict
			}
			snap, e := tx.Get(parent.Collection("spaces").Doc(old.ID))
			result, e = decode(snap, e, r.ID)
			if e == nil && result.CreatorID != actor {
				return feature.ErrUnavailable
			}
			return e
		}
		if status.Code(e) != codes.NotFound {
			return e
		}
		result = feature.Space{ID: id, RepositoryID: r.ID, Name: name, CreatorID: actor, Version: 1, CreatedAt: now, UpdatedAt: now}
		if e = tx.Create(ref, encode(result)); e != nil {
			return e
		}
		return tx.Create(rec, receipt{1, id, payload})
	})
	if err != nil {
		return feature.Space{}, err
	}
	return result, nil
}
