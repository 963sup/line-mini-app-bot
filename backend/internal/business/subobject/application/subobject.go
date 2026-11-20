package application

import (
	"context"
	"errors"
	"regexp"
	"strings"

	objectDomainEntity "github.com/963sup/line-mini-app-bot/internal/business/object/domain/entity"
	repositoryDomainEntity "github.com/963sup/line-mini-app-bot/internal/business/repository/domain/entity"
	subobjectApplicationDto "github.com/963sup/line-mini-app-bot/internal/business/subobject/application/dto"
	subobjectApplicationPort "github.com/963sup/line-mini-app-bot/internal/business/subobject/application/port"
	subobjectDomainEntity "github.com/963sup/line-mini-app-bot/internal/business/subobject/domain/entity"
)

var keyPattern = regexp.MustCompile(`^[A-Za-z0-9_-]{16,128}$`)

type Service struct {
	Repositories subobjectApplicationPort.Repositories
	Store        subobjectApplicationPort.Store
}

func (s *Service) scope(ctx context.Context, actor, repositoryID, parentObjectID string) (repositoryDomainEntity.Repository, error) {
	if s == nil || s.Store == nil || s.Repositories == nil {
		return repositoryDomainEntity.Repository{}, subobjectDomainEntity.ErrUnavailable
	}
	if actor == "" || !repositoryDomainEntity.ValidID(repositoryID) || !repositoryDomainEntity.ValidID(parentObjectID) {
		return repositoryDomainEntity.Repository{}, subobjectDomainEntity.ErrNotFound
	}
	r, e := s.Repositories.Get(ctx, actor, repositoryID)
	if errors.Is(e, repositoryDomainEntity.ErrNotFound) {
		return repositoryDomainEntity.Repository{}, subobjectDomainEntity.ErrNotFound
	}
	if e != nil {
		return repositoryDomainEntity.Repository{}, e
	}
	if r.ID != repositoryID || r.Status != "ACTIVE" || !repositoryDomainEntity.ValidOwner(r.OwnerID) {
		return repositoryDomainEntity.Repository{}, subobjectDomainEntity.ErrNotFound
	}
	return r, nil
}

func (s *Service) List(ctx context.Context, actor, repositoryID, parentObjectID, cursor string) (subobjectApplicationDto.Page, error) {
	r, e := s.scope(ctx, actor, repositoryID, parentObjectID)
	if e != nil {
		return subobjectApplicationDto.Page{}, e
	}
	if len(cursor) > 1024 {
		return subobjectApplicationDto.Page{}, subobjectDomainEntity.ErrInput
	}
	return s.Store.List(ctx, r, actor, parentObjectID, cursor)
}

func (s *Service) Get(ctx context.Context, actor, repositoryID, parentObjectID, subobjectID string) (subobjectDomainEntity.Subobject, error) {
	r, e := s.scope(ctx, actor, repositoryID, parentObjectID)
	if e != nil {
		return subobjectDomainEntity.Subobject{}, e
	}
	if !repositoryDomainEntity.ValidID(subobjectID) {
		return subobjectDomainEntity.Subobject{}, subobjectDomainEntity.ErrNotFound
	}
	return s.Store.Get(ctx, r, actor, parentObjectID, subobjectID)
}

func (s *Service) Create(ctx context.Context, actor, repositoryID, parentObjectID, key string, input subobjectDomainEntity.Content) (subobjectDomainEntity.Subobject, error) {
	r, e := s.scope(ctx, actor, repositoryID, parentObjectID)
	if e != nil {
		return subobjectDomainEntity.Subobject{}, e
	}
	if !keyPattern.MatchString(key) {
		return subobjectDomainEntity.Subobject{}, subobjectDomainEntity.ErrInput
	}
	if strings.TrimSpace(input.Type) == objectDomainEntity.TypeWork && input.CompletedQuantity == "" {
		input.CompletedQuantity = "0"
	}
	c, e := subobjectDomainEntity.Normalize(input)
	if e != nil {
		return subobjectDomainEntity.Subobject{}, e
	}
	return s.Store.Create(ctx, r, actor, parentObjectID, key, c)
}

func (s *Service) Update(ctx context.Context, actor, repositoryID, parentObjectID, subobjectID string, input subobjectDomainEntity.Update) (subobjectDomainEntity.Subobject, error) {
	r, e := s.scope(ctx, actor, repositoryID, parentObjectID)
	if e != nil {
		return subobjectDomainEntity.Subobject{}, e
	}
	if !repositoryDomainEntity.ValidID(subobjectID) {
		return subobjectDomainEntity.Subobject{}, subobjectDomainEntity.ErrNotFound
	}
	return s.Store.Update(ctx, r, actor, parentObjectID, subobjectID, input)
}
