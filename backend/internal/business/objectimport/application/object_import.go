package application

import (
	"context"
	"errors"
	"regexp"
	"time"

	objectimportApplicationPort "github.com/963sup/line-mini-app-bot/internal/business/objectimport/application/port"
	objectimportDomainEntity "github.com/963sup/line-mini-app-bot/internal/business/objectimport/domain/entity"
	"github.com/963sup/line-mini-app-bot/internal/business/repository"
)

var keyPattern = regexp.MustCompile(`^[A-Za-z0-9_-]{16,128}$`)

type Service struct {
	Repositories objectimportApplicationPort.Repositories
	Objects      objectimportApplicationPort.Objects
	Store        objectimportApplicationPort.Store
}

func (s *Service) scope(ctx context.Context, actor, rid string) (repository.Repository, error) {
	if s == nil || s.Store == nil || s.Repositories == nil {
		return repository.Repository{}, objectimportDomainEntity.ErrUnavailable
	}
	if actor == "" || !repository.ValidID(rid) {
		return repository.Repository{}, objectimportDomainEntity.ErrNotFound
	}
	r, e := s.Repositories.Get(ctx, actor, rid)
	if errors.Is(e, repository.ErrNotFound) {
		return repository.Repository{}, objectimportDomainEntity.ErrNotFound
	}
	if e != nil {
		return repository.Repository{}, objectimportDomainEntity.ErrUnavailable
	}
	if r.ID != rid || r.Status != "ACTIVE" || !repository.ValidOwner(r.OwnerID) {
		return repository.Repository{}, objectimportDomainEntity.ErrNotFound
	}
	return r, nil
}

func (s *Service) Create(ctx context.Context, actor, rid, key string, input objectimportDomainEntity.Creation) (objectimportDomainEntity.Job, error) {
	r, e := s.scope(ctx, actor, rid)
	if e != nil {
		return objectimportDomainEntity.Job{}, e
	}
	if !keyPattern.MatchString(key) {
		return objectimportDomainEntity.Job{}, objectimportDomainEntity.ErrInput
	}
	j, payload, e := objectimportDomainEntity.Prepare(input)
	if e != nil {
		return objectimportDomainEntity.Job{}, e
	}
	j.ID, e = objectimportDomainEntity.NewID()
	if e != nil {
		return objectimportDomainEntity.Job{}, objectimportDomainEntity.ErrUnavailable
	}
	j.RepositoryID = rid
	j.CreatorID = actor
	j.CreatedAt = time.Now().UTC().Truncate(time.Microsecond)
	j.UpdatedAt = j.CreatedAt
	for i := range j.Rows {
		j.Rows[i].ID, e = objectimportDomainEntity.NewID()
		if e != nil {
			return objectimportDomainEntity.Job{}, objectimportDomainEntity.ErrUnavailable
		}
	}
	if !objectimportDomainEntity.WithinSizeLimit(j) {
		return objectimportDomainEntity.Job{}, objectimportDomainEntity.ErrInput
	}
	return s.Store.Create(ctx, r, actor, key, payload, j)
}

func (s *Service) Get(ctx context.Context, actor, rid, id string) (objectimportDomainEntity.Job, error) {
	r, e := s.scope(ctx, actor, rid)
	if e != nil {
		return objectimportDomainEntity.Job{}, e
	}
	if !repository.ValidID(id) {
		return objectimportDomainEntity.Job{}, objectimportDomainEntity.ErrNotFound
	}
	return s.Store.Get(ctx, r, actor, id)
}

func (s *Service) Review(ctx context.Context, actor, rid, id string, input objectimportDomainEntity.Review) (objectimportDomainEntity.Job, error) {
	r, e := s.scope(ctx, actor, rid)
	if e != nil {
		return objectimportDomainEntity.Job{}, e
	}
	if !repository.ValidID(id) {
		return objectimportDomainEntity.Job{}, objectimportDomainEntity.ErrNotFound
	}
	return s.Store.Review(ctx, r, actor, id, input)
}

func (s *Service) Confirm(ctx context.Context, actor, rid, id string, expected int64) (objectimportDomainEntity.Job, error) {
	r, e := s.scope(ctx, actor, rid)
	if e != nil {
		return objectimportDomainEntity.Job{}, e
	}
	if !repository.ValidID(id) {
		return objectimportDomainEntity.Job{}, objectimportDomainEntity.ErrNotFound
	}
	if s.Objects == nil {
		return objectimportDomainEntity.Job{}, objectimportDomainEntity.ErrUnavailable
	}
	j, e := s.Store.BeginConfirm(ctx, r, actor, id, expected)
	if e != nil {
		return objectimportDomainEntity.Job{}, e
	}
	for _, row := range j.Rows {
		if row.ResultObjectID != "" {
			continue
		}
		key := objectimportDomainEntity.Hash([]string{"object-import-v1", j.ID, row.ID})
		created, e := s.Objects.Create(ctx, actor, rid, key, objectimportDomainEntity.ObjectContent(j, row))
		if e != nil {
			return j, objectimportDomainEntity.ErrUnavailable
		}
		if !repository.ValidID(created.ID) {
			return j, objectimportDomainEntity.ErrUnavailable
		}
		j, e = s.Store.RecordResult(ctx, r, actor, id, row.ID, created.ID, expected)
		if e != nil {
			return objectimportDomainEntity.Job{}, e
		}
	}
	return j, nil
}
