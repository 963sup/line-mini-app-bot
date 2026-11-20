package firestore

import (
	objectimportApplicationPort "github.com/963sup/line-mini-app-bot/internal/business/objectimport/application/port"
	objectimportDomainEntity "github.com/963sup/line-mini-app-bot/internal/business/objectimport/domain/entity"
)

import (
	"context"
	"encoding/json"
	"errors"
	"time"

	db "cloud.google.com/go/firestore"

	"github.com/963sup/line-mini-app-bot/internal/business/repository"
	"github.com/963sup/line-mini-app-bot/internal/persistence/grpcstatus"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type Store struct {
	Client    *db.Client
	ChannelID string
}

var _ objectimportApplicationPort.Store = (*Store)(nil)

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

func encode(j objectimportDomainEntity.Job) (document, error) {
	if e := objectimportDomainEntity.ValidateStored(j); e != nil {
		return document{}, e
	}
	b, e := json.Marshal(toStoredJob(j))
	if e != nil || len(b) > objectimportDomainEntity.MaxDraftBytes {
		return document{}, objectimportDomainEntity.ErrInput
	}
	return document{1, string(b)}, nil
}

func decode(snap *db.DocumentSnapshot, e error, rid, actor string) (objectimportDomainEntity.Job, error) {
	if status.Code(e) == codes.NotFound {
		return objectimportDomainEntity.Job{}, objectimportDomainEntity.ErrNotFound
	}
	if e != nil {
		return objectimportDomainEntity.Job{}, e
	}
	var d document
	if e = snap.DataTo(&d); e != nil {
		return objectimportDomainEntity.Job{}, e
	}
	var stored storedJob
	if d.SchemaVersion != 1 || len(d.Payload) > objectimportDomainEntity.MaxDraftBytes || json.Unmarshal([]byte(d.Payload), &stored) != nil {
		return objectimportDomainEntity.Job{}, objectimportDomainEntity.ErrUnavailable
	}
	j := fromStoredJob(stored)
	if j.ID != snap.Ref.ID || j.RepositoryID != rid || objectimportDomainEntity.ValidateStored(j) != nil {
		return objectimportDomainEntity.Job{}, objectimportDomainEntity.ErrUnavailable
	}
	if j.CreatorID != actor {
		return objectimportDomainEntity.Job{}, objectimportDomainEntity.ErrNotFound
	}
	return j, nil
}
func (s *Store) ready() bool { return s != nil && s.Client != nil && s.ChannelID != "" }
func (s *Store) repositoryRef(rid string) *db.DocumentRef {
	return s.Client.Collection("channels").Doc(s.ChannelID).Collection("repositories").Doc(rid)
}

func checkActiveRepository(tx *db.Transaction, parent *db.DocumentRef, r repository.Repository) error {
	snap, e := tx.Get(parent)
	if status.Code(e) == codes.NotFound {
		return objectimportDomainEntity.ErrNotFound
	}
	if e != nil {
		return e
	}
	var c struct {
		ID      string `firestore:"id"`
		OwnerID string `firestore:"ownerId"`
		Status  string `firestore:"status"`
	}
	if e = snap.DataTo(&c); e != nil {
		return e
	}
	if !r.MatchesActive(c.ID, c.OwnerID, c.Status) {
		return objectimportDomainEntity.ErrNotFound
	}
	return nil
}

func translate(e error) error {
	if e == nil {
		return nil
	}
	for _, v := range []error{objectimportDomainEntity.ErrInput, objectimportDomainEntity.ErrNotFound, objectimportDomainEntity.ErrForbidden, objectimportDomainEntity.ErrConflict, objectimportDomainEntity.ErrCreationConflict, objectimportDomainEntity.ErrUnavailable, context.Canceled, context.DeadlineExceeded} {
		if errors.Is(e, v) {
			return e
		}
	}
	return grpcstatus.NormalizeContextError(e)
}

func (s *Store) Create(ctx context.Context, r repository.Repository, actor, key, payload string, j objectimportDomainEntity.Job) (result objectimportDomainEntity.Job, err error) {
	defer func() { err = translate(err) }()
	if !s.ready() {
		return result, objectimportDomainEntity.ErrUnavailable
	}
	if j.RepositoryID != r.ID || j.CreatorID != actor {
		return result, objectimportDomainEntity.ErrInput
	}
	d, e := encode(j)
	if e != nil {
		return result, e
	}
	parent := s.repositoryRef(r.ID)
	rec := parent.Collection("objectImportCreationReceipts").Doc(objectimportDomainEntity.Hash([]string{"object-import-create-v1", actor, key}))
	err = s.Client.RunTransaction(ctx, func(_ context.Context, tx *db.Transaction) error {
		result = objectimportDomainEntity.Job{}
		if e := checkActiveRepository(tx, parent, r); e != nil {
			return e
		}
		snap, e := tx.Get(rec)
		if e == nil {
			var old receipt
			if e = snap.DataTo(&old); e != nil {
				return e
			}
			if old.SchemaVersion != 1 || !repository.ValidID(old.ID) || len(old.PayloadHash) != 64 {
				return objectimportDomainEntity.ErrUnavailable
			}
			if old.PayloadHash != payload {
				return objectimportDomainEntity.ErrCreationConflict
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
		return objectimportDomainEntity.Job{}, err
	}
	return result, nil
}

func (s *Store) Get(ctx context.Context, r repository.Repository, actor, id string) (result objectimportDomainEntity.Job, err error) {
	return s.mutate(ctx, r, actor, id, func(j objectimportDomainEntity.Job) (objectimportDomainEntity.Job, error) { return j, nil })
}

func (s *Store) mutate(ctx context.Context, r repository.Repository, actor, id string, apply func(objectimportDomainEntity.Job) (objectimportDomainEntity.Job, error)) (result objectimportDomainEntity.Job, err error) {
	defer func() { err = translate(err) }()
	if !s.ready() {
		return result, objectimportDomainEntity.ErrUnavailable
	}
	parent := s.repositoryRef(r.ID)
	ref := parent.Collection("objectImports").Doc(id)
	err = s.Client.RunTransaction(ctx, func(_ context.Context, tx *db.Transaction) error {
		result = objectimportDomainEntity.Job{}
		if e := checkActiveRepository(tx, parent, r); e != nil {
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
		return objectimportDomainEntity.Job{}, err
	}
	return result, nil
}

func nowAfter(j objectimportDomainEntity.Job) time.Time {
	now := time.Now().UTC().Truncate(time.Microsecond)
	if !now.After(j.UpdatedAt) {
		now = j.UpdatedAt.Add(time.Microsecond)
	}
	return now
}

func (s *Store) Review(ctx context.Context, r repository.Repository, actor, id string, input objectimportDomainEntity.Review) (objectimportDomainEntity.Job, error) {
	return s.mutate(ctx, r, actor, id, func(j objectimportDomainEntity.Job) (objectimportDomainEntity.Job, error) {
		return objectimportDomainEntity.ApplyReview(j, actor, input, nowAfter(j))
	})
}

func (s *Store) BeginConfirm(ctx context.Context, r repository.Repository, actor, id string, version int64) (objectimportDomainEntity.Job, error) {
	return s.mutate(ctx, r, actor, id, func(j objectimportDomainEntity.Job) (objectimportDomainEntity.Job, error) {
		return objectimportDomainEntity.Begin(j, actor, version, nowAfter(j))
	})
}

func (s *Store) RecordResult(ctx context.Context, r repository.Repository, actor, id, rowID, objectID string, reviewVersion int64) (objectimportDomainEntity.Job, error) {
	return s.mutate(ctx, r, actor, id, func(j objectimportDomainEntity.Job) (objectimportDomainEntity.Job, error) {
		return objectimportDomainEntity.Record(j, actor, rowID, objectID, reviewVersion, nowAfter(j))
	})
}
