package firestore

import (
	"context"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"errors"
	"strconv"
	"time"

	db "cloud.google.com/go/firestore"
	"github.com/963sup/line-mini-app-bot/internal/business/repository"
	feature "github.com/963sup/line-mini-app-bot/internal/business/subobject"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type Store struct {
	Client    *db.Client
	ChannelID string
}

var _ feature.Store = (*Store)(nil)

type document struct {
	SchemaVersion     int64     `firestore:"schemaVersion"`
	ID                string    `firestore:"id"`
	RepositoryID      string    `firestore:"repositoryId"`
	ParentObjectID    string    `firestore:"parentObjectId"`
	CreatorID         string    `firestore:"creatorId"`
	Name              string    `firestore:"name"`
	Type              string    `firestore:"type"`
	Quantity          string    `firestore:"quantity"`
	Unit              string    `firestore:"unit"`
	CompletedQuantity string    `firestore:"completedQuantity,omitempty"`
	Status            string    `firestore:"status"`
	Version           int64     `firestore:"version"`
	CreatedAt         time.Time `firestore:"createdAt"`
	UpdatedAt         time.Time `firestore:"updatedAt"`
}

func encode(j feature.Subobject) document {
	return document{SchemaVersion: 1, ID: j.ID, RepositoryID: j.RepositoryID, ParentObjectID: j.ParentObjectID, CreatorID: j.CreatorID, Name: j.Name, Type: j.Type, Quantity: j.Quantity, Unit: j.Unit, CompletedQuantity: j.CompletedQuantity, Status: j.Status, Version: j.Version, CreatedAt: j.CreatedAt, UpdatedAt: j.UpdatedAt}
}

func decode(snap *db.DocumentSnapshot, e error, rid, pid, owner string) (feature.Subobject, error) {
	if status.Code(e) == codes.NotFound {
		return feature.Subobject{}, feature.ErrNotFound
	}
	if e != nil {
		return feature.Subobject{}, e
	}
	var d document
	if e = snap.DataTo(&d); e != nil {
		return feature.Subobject{}, e
	}
	c := feature.Content{Name: d.Name, Type: d.Type, Quantity: d.Quantity, Unit: d.Unit, CompletedQuantity: d.CompletedQuantity}
	normalized, e := feature.Normalize(c)
	derived, se := feature.Status(c)
	if e != nil || se != nil || normalized != c || derived != d.Status || d.SchemaVersion != 1 || d.ID != snap.Ref.ID || !repository.ValidID(d.ID) || d.RepositoryID != rid || d.ParentObjectID != pid || d.CreatorID != owner || owner == "" || d.Version < 1 || d.Version > repository.MaxResourceNumber || d.CreatedAt.IsZero() || d.UpdatedAt.Before(d.CreatedAt) {
		return feature.Subobject{}, feature.ErrUnavailable
	}
	return feature.Subobject{Content: c, ID: d.ID, RepositoryID: rid, ParentObjectID: pid, CreatorID: d.CreatorID, Status: d.Status, Version: d.Version, CreatedAt: d.CreatedAt, UpdatedAt: d.UpdatedAt}, nil
}
func (s *Store) ready() bool { return s != nil && s.Client != nil && s.ChannelID != "" }
func (s *Store) repository(rid string) *db.DocumentRef {
	return s.Client.Collection("channels").Doc(s.ChannelID).Collection("repositories").Doc(rid)
}

func (s *Store) parent(rid, pid string) *db.DocumentRef {
	return s.repository(rid).Collection("objects").Doc(pid)
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
	for _, known := range []error{feature.ErrInput, feature.ErrNotFound, feature.ErrForbidden, feature.ErrConflict, feature.ErrCreationConflict, feature.ErrProgressRemoval, feature.ErrUnavailable, context.Canceled, context.DeadlineExceeded} {
		if errors.Is(e, known) {
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

// scope reads only authority/scope fields; no parent quantity, progress or version is written.
func (s *Store) scope(tx *db.Transaction, r repository.Repository, pid, actor string, write bool) (string, error) {
	snap, e := tx.Get(s.repository(r.ID))
	if status.Code(e) == codes.NotFound {
		return "", feature.ErrNotFound
	}
	if e != nil {
		return "", e
	}
	var repo struct {
		ID      string `firestore:"id"`
		OwnerID string `firestore:"ownerId"`
		Status  string `firestore:"status"`
	}
	if e = snap.DataTo(&repo); e != nil {
		return "", e
	}
	if repo.ID != r.ID || repo.OwnerID != r.OwnerID || repo.Status != "ACTIVE" {
		return "", feature.ErrNotFound
	}
	snap, e = tx.Get(s.parent(r.ID, pid))
	if status.Code(e) == codes.NotFound {
		return "", feature.ErrNotFound
	}
	if e != nil {
		return "", e
	}
	var parent struct {
		ID            string `firestore:"id"`
		RepositoryID  string `firestore:"repositoryId"`
		CreatorID     string `firestore:"creatorId"`
		SchemaVersion int64  `firestore:"schemaVersion"`
	}
	if e = snap.DataTo(&parent); e != nil {
		return "", e
	}
	if parent.ID != pid || parent.RepositoryID != r.ID || parent.CreatorID == "" || parent.SchemaVersion != 1 {
		return "", feature.ErrNotFound
	}
	if write && parent.CreatorID != actor {
		return "", feature.ErrForbidden
	}
	return parent.CreatorID, nil
}

func (s *Store) Get(ctx context.Context, r repository.Repository, actor, pid, id string) (result feature.Subobject, err error) {
	defer func() { err = translate(err) }()
	if !s.ready() {
		return result, feature.ErrUnavailable
	}
	err = s.Client.RunTransaction(ctx, func(_ context.Context, tx *db.Transaction) error {
		result = feature.Subobject{}
		owner, e := s.scope(tx, r, pid, actor, false)
		if e != nil {
			return e
		}
		snap, e := tx.Get(s.parent(r.ID, pid).Collection("subobjects").Doc(id))
		result, e = decode(snap, e, r.ID, pid, owner)
		return e
	})
	if err != nil {
		return feature.Subobject{}, err
	}
	return result, nil
}

type cursor struct{ Scope, ID string }

func (s *Store) List(ctx context.Context, r repository.Repository, actor, pid, token string) (result feature.Page, err error) {
	defer func() { err = translate(err) }()
	if !s.ready() {
		return result, feature.ErrUnavailable
	}
	scope := hash([]string{"subobject-cursor-v1", s.ChannelID, r.ID, pid, actor})
	c := cursor{}
	if token != "" {
		b, e := base64.RawURLEncoding.DecodeString(token)
		if e != nil || json.Unmarshal(b, &c) != nil || c.Scope != scope || !repository.ValidID(c.ID) {
			return result, feature.ErrInput
		}
	}
	err = s.Client.RunTransaction(ctx, func(_ context.Context, tx *db.Transaction) error {
		result = feature.Page{Items: []feature.Subobject{}}
		owner, e := s.scope(tx, r, pid, actor, false)
		if e != nil {
			return e
		}
		q := s.parent(r.ID, pid).Collection("subobjects").OrderBy(db.DocumentID, db.Asc).Limit(21)
		if c.ID != "" {
			q = q.StartAfter(c.ID)
		}
		docs, e := tx.Documents(q).GetAll()
		if e != nil {
			return e
		}
		more := len(docs) > 20
		if more {
			docs = docs[:20]
		}
		for _, snap := range docs {
			j, e := decode(snap, nil, r.ID, pid, owner)
			if e != nil {
				return e
			}
			result.Items = append(result.Items, j)
		}
		if more {
			b, _ := json.Marshal(cursor{Scope: scope, ID: docs[len(docs)-1].Ref.ID})
			result.NextCursor = base64.RawURLEncoding.EncodeToString(b)
		}
		return nil
	})
	if err != nil {
		return feature.Page{}, err
	}
	return result, nil
}

type receipt struct {
	SchemaVersion int64  `firestore:"schemaVersion"`
	ID            string `firestore:"id"`
	PayloadHash   string `firestore:"payloadHash"`
}

func (s *Store) Create(ctx context.Context, r repository.Repository, actor, pid, key string, c feature.Content) (result feature.Subobject, err error) {
	defer func() { err = translate(err) }()
	if !s.ready() {
		return result, feature.ErrUnavailable
	}
	parent := s.parent(r.ID, pid)
	ref := parent.Collection("subobjects").NewDoc() // IDs follow the same opaque 32-hex convention as Object.
	id := hash([]string{"subobject-id-v1", ref.ID})[:32]
	ref = parent.Collection("subobjects").Doc(id)
	rec := parent.Collection("subobjectCreationReceipts").Doc(hash([]string{"subobject-create-v1", actor, pid, key}))
	payload := hash(struct {
		Version                                       int
		Name, Type, Quantity, Unit, CompletedQuantity string
	}{1, c.Name, c.Type, c.Quantity, c.Unit, c.CompletedQuantity})
	now := time.Now().UTC().Truncate(time.Microsecond)
	err = s.Client.RunTransaction(ctx, func(_ context.Context, tx *db.Transaction) error {
		result = feature.Subobject{}
		owner, e := s.scope(tx, r, pid, actor, true)
		if e != nil {
			return e
		}
		snap, e := tx.Get(rec)
		if e == nil {
			var old receipt
			if e = snap.DataTo(&old); e != nil {
				return e
			}
			if old.SchemaVersion != 1 || !repository.ValidID(old.ID) || len(old.PayloadHash) != 64 {
				return feature.ErrUnavailable
			}
			if old.PayloadHash != payload {
				return feature.ErrCreationConflict
			}
			snap, e = tx.Get(parent.Collection("subobjects").Doc(old.ID))
			result, e = decode(snap, e, r.ID, pid, owner)
			return e
		}
		if status.Code(e) != codes.NotFound {
			return e
		}
		derived, e := feature.Status(c)
		if e != nil {
			return e
		}
		result = feature.Subobject{Content: c, ID: id, RepositoryID: r.ID, ParentObjectID: pid, CreatorID: owner, Status: derived, Version: 1, CreatedAt: now, UpdatedAt: now}
		if e = tx.Create(ref, encode(result)); e != nil {
			return e
		}
		return tx.Create(rec, receipt{1, id, payload})
	})
	if err != nil {
		return feature.Subobject{}, err
	}
	return result, nil
}

func (s *Store) Update(ctx context.Context, r repository.Repository, actor, pid, id string, input feature.Update) (result feature.Subobject, err error) {
	defer func() { err = translate(err) }()
	if !s.ready() {
		return result, feature.ErrUnavailable
	}
	ref := s.parent(r.ID, pid).Collection("subobjects").Doc(id)
	err = s.Client.RunTransaction(ctx, func(_ context.Context, tx *db.Transaction) error {
		result = feature.Subobject{}
		owner, e := s.scope(tx, r, pid, actor, true)
		if e != nil {
			return e
		}
		snap, e := tx.Get(ref)
		old, e := decode(snap, e, r.ID, pid, owner)
		if e != nil {
			return e
		}
		result, e = feature.ApplyUpdate(old, actor, input)
		if e != nil {
			return e
		}
		if result.Version == old.Version {
			return nil
		}
		result.UpdatedAt = time.Now().UTC().Truncate(time.Microsecond)
		if !result.UpdatedAt.After(old.UpdatedAt) {
			result.UpdatedAt = old.UpdatedAt.Add(time.Microsecond)
		}
		if e = tx.Set(ref, encode(result)); e != nil {
			return e
		}
		return tx.Create(ref.Collection("changes").Doc(strconv.FormatInt(result.Version, 10)), map[string]any{"actorId": actor, "version": result.Version, "at": result.UpdatedAt, "before": encode(old), "after": encode(result)})
	})
	if err != nil {
		return feature.Subobject{}, err
	}
	return result, nil
}
