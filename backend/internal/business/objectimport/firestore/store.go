package firestore

import (
	"context"
	"encoding/json"
	"errors"
	"time"

	db "cloud.google.com/go/firestore"
	feature "github.com/963sup/line-mini-app-bot/internal/business/objectimport"
	"github.com/963sup/line-mini-app-bot/internal/business/repository"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type Store struct {
	Client    *db.Client
	ChannelID string
}

var _ feature.Store = (*Store)(nil)

// JSON payload provides explicit adapter serialization and a conservative bounded document.
type document struct {
	SchemaVersion int64  `firestore:"schemaVersion"`
	Payload       string `firestore:"payload"`
}
type receipt struct {
	SchemaVersion int64  `firestore:"schemaVersion"`
	ID            string `firestore:"id"`
	PayloadHash   string `firestore:"payloadHash"`
}

func encode(j feature.Job) (document, error) {
	if e := feature.ValidateStored(j); e != nil {
		return document{}, e
	}
	b, e := json.Marshal(toStored(j))
	if e != nil || len(b) > feature.MaxDraftBytes {
		return document{}, feature.ErrInput
	}
	return document{1, string(b)}, nil
}

func decode(snap *db.DocumentSnapshot, e error, rid, actor string) (feature.Job, error) {
	if status.Code(e) == codes.NotFound {
		return feature.Job{}, feature.ErrNotFound
	}
	if e != nil {
		return feature.Job{}, e
	}
	var d document
	if e = snap.DataTo(&d); e != nil {
		return feature.Job{}, e
	}
	var stored storedJob
	if d.SchemaVersion != 1 || len(d.Payload) > feature.MaxDraftBytes || json.Unmarshal([]byte(d.Payload), &stored) != nil {
		return feature.Job{}, feature.ErrUnavailable
	}
	j := fromStored(stored)
	if j.ID != snap.Ref.ID || j.RepositoryID != rid || feature.ValidateStored(j) != nil {
		return feature.Job{}, feature.ErrUnavailable
	}
	if j.CreatorID != actor {
		return feature.Job{}, feature.ErrNotFound
	}
	return j, nil
}
func (s *Store) ready() bool { return s != nil && s.Client != nil && s.ChannelID != "" }
func (s *Store) parent(rid string) *db.DocumentRef {
	return s.Client.Collection("channels").Doc(s.ChannelID).Collection("repositories").Doc(rid)
}

func container(tx *db.Transaction, parent *db.DocumentRef, r repository.Repository) error {
	snap, e := tx.Get(parent)
	if status.Code(e) == codes.NotFound {
		return feature.ErrNotFound
	}
	if e != nil {
		return e
	}
	var c struct {
		ID      string `firestore:"id"`
		GroupID string `firestore:"groupId"`
		Status  string `firestore:"status"`
	}
	if e = snap.DataTo(&c); e != nil {
		return e
	}
	if c.ID != r.ID || c.GroupID != r.GroupID || c.Status != "ACTIVE" {
		return feature.ErrNotFound
	}
	return nil
}

type failure struct{ cause error }

func (e *failure) Error() string { return "object import storage unavailable" }
func (e *failure) Unwrap() error { return feature.ErrUnavailable }
func translate(e error) error {
	if e == nil {
		return nil
	}
	for _, v := range []error{feature.ErrInput, feature.ErrNotFound, feature.ErrForbidden, feature.ErrConflict, feature.ErrCreationConflict, feature.ErrUnavailable, context.Canceled, context.DeadlineExceeded} {
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
	return &failure{e}
}

func (s *Store) Create(ctx context.Context, r repository.Repository, actor, key, payload string, j feature.Job) (result feature.Job, err error) {
	defer func() { err = translate(err) }()
	if !s.ready() {
		return result, feature.ErrUnavailable
	}
	if j.RepositoryID != r.ID || j.CreatorID != actor {
		return result, feature.ErrInput
	}
	d, e := encode(j)
	if e != nil {
		return result, e
	}
	parent := s.parent(r.ID)
	rec := parent.Collection("objectImportCreationReceipts").Doc(feature.Hash([]string{"object-import-create-v1", actor, key}))
	err = s.Client.RunTransaction(ctx, func(_ context.Context, tx *db.Transaction) error {
		result = feature.Job{}
		if e := container(tx, parent, r); e != nil {
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
			snap, e = tx.Get(parent.Collection("objectImports").Doc(old.ID))
			result, e = decode(snap, e, r.ID, actor)
			return e
		}
		if status.Code(e) != codes.NotFound {
			return e
		}
		if e = tx.Create(parent.Collection("objectImports").Doc(j.ID), d); e != nil {
			return e
		}
		if e = tx.Create(rec, receipt{1, j.ID, payload}); e != nil {
			return e
		}
		result = j
		return nil
	})
	if err != nil {
		return feature.Job{}, err
	}
	return result, nil
}

func (s *Store) Get(ctx context.Context, r repository.Repository, actor, id string) (result feature.Job, err error) {
	return s.mutate(ctx, r, actor, id, func(j feature.Job) (feature.Job, error) { return j, nil })
}

func (s *Store) mutate(ctx context.Context, r repository.Repository, actor, id string, apply func(feature.Job) (feature.Job, error)) (result feature.Job, err error) {
	defer func() { err = translate(err) }()
	if !s.ready() {
		return result, feature.ErrUnavailable
	}
	parent := s.parent(r.ID)
	ref := parent.Collection("objectImports").Doc(id)
	err = s.Client.RunTransaction(ctx, func(_ context.Context, tx *db.Transaction) error {
		result = feature.Job{}
		if e := container(tx, parent, r); e != nil {
			return e
		}
		snap, e := tx.Get(ref)
		old, e := decode(snap, e, r.ID, actor)
		if e != nil {
			return e
		}
		result, e = apply(old)
		if e != nil {
			return e
		}
		if result.Version == old.Version {
			return nil
		}
		d, e := encode(result)
		if e != nil {
			return e
		}
		return tx.Set(ref, d)
	})
	if err != nil {
		return feature.Job{}, err
	}
	return result, nil
}

func nowAfter(j feature.Job) time.Time {
	now := time.Now().UTC().Truncate(time.Microsecond)
	if !now.After(j.UpdatedAt) {
		now = j.UpdatedAt.Add(time.Microsecond)
	}
	return now
}

func (s *Store) Review(ctx context.Context, r repository.Repository, actor, id string, input feature.Review) (feature.Job, error) {
	return s.mutate(ctx, r, actor, id, func(j feature.Job) (feature.Job, error) { return feature.ApplyReview(j, actor, input, nowAfter(j)) })
}

func (s *Store) BeginConfirm(ctx context.Context, r repository.Repository, actor, id string, version int64) (feature.Job, error) {
	return s.mutate(ctx, r, actor, id, func(j feature.Job) (feature.Job, error) { return feature.Begin(j, actor, version, nowAfter(j)) })
}

func (s *Store) RecordResult(ctx context.Context, r repository.Repository, actor, id, rowID, objectID string, reviewVersion int64) (feature.Job, error) {
	return s.mutate(ctx, r, actor, id, func(j feature.Job) (feature.Job, error) {
		return feature.Record(j, actor, rowID, objectID, reviewVersion, nowAfter(j))
	})
}
