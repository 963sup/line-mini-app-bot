package space

import (
	"context"
	"errors"
	"regexp"

	repositoryApplicationDtoSpace "github.com/963sup/line-mini-app-bot/internal/business/repository/application/dto/space"
	repositoryApplicationPortSpace "github.com/963sup/line-mini-app-bot/internal/business/repository/application/port/space"
	repositoryDomainEntity "github.com/963sup/line-mini-app-bot/internal/business/repository/domain/entity"
	repositoryDomainEntitySpace "github.com/963sup/line-mini-app-bot/internal/business/repository/domain/entity/space"
)

var creationKey = regexp.MustCompile(`^[A-Za-z0-9_-]{16,128}$`)

type Service struct {
	Repositories repositoryApplicationPortSpace.Repositories
	Store        repositoryApplicationPortSpace.Store
}

func (s *Service) scope(ctx context.Context, actor, rid string) (repositoryDomainEntity.Repository, error) {
	if s == nil || s.Repositories == nil || s.Store == nil {
		return repositoryDomainEntity.Repository{}, repositoryDomainEntitySpace.ErrUnavailable
	}
	if actor == "" || !repositoryDomainEntity.ValidID(rid) {
		return repositoryDomainEntity.Repository{}, repositoryDomainEntitySpace.ErrNotFound
	}
	r, e := s.Repositories.Get(ctx, actor, rid)
	if errors.Is(e, repositoryDomainEntity.ErrNotFound) {
		return repositoryDomainEntity.Repository{}, repositoryDomainEntitySpace.ErrNotFound
	}
	if e != nil {
		return repositoryDomainEntity.Repository{}, e
	}
	if r.ID != rid || r.Status != "ACTIVE" || !repositoryDomainEntity.ValidTypedOwner(r.OwnerType, r.OwnerID) {
		return repositoryDomainEntity.Repository{}, repositoryDomainEntitySpace.ErrNotFound
	}
	return r, nil
}

func (s *Service) List(ctx context.Context, actor, rid, cursor string) (repositoryApplicationDtoSpace.Page, error) {
	r, e := s.scope(ctx, actor, rid)
	if e != nil {
		return repositoryApplicationDtoSpace.Page{}, e
	}
	if len(cursor) > 1024 {
		return repositoryApplicationDtoSpace.Page{}, repositoryDomainEntitySpace.ErrInput
	}
	return s.Store.List(ctx, r, actor, cursor)
}

func (s *Service) Get(ctx context.Context, actor, rid, id string) (repositoryDomainEntitySpace.Space, error) {
	r, e := s.scope(ctx, actor, rid)
	if e != nil {
		return repositoryDomainEntitySpace.Space{}, e
	}
	if !repositoryDomainEntity.ValidID(id) {
		return repositoryDomainEntitySpace.Space{}, repositoryDomainEntitySpace.ErrNotFound
	}
	return s.Store.Get(ctx, r, id)
}

func (s *Service) Create(ctx context.Context, actor, rid, key, name string) (repositoryDomainEntitySpace.Space, error) {
	r, e := s.scope(ctx, actor, rid)
	if e != nil {
		return repositoryDomainEntitySpace.Space{}, e
	}
	name, e = repositoryDomainEntitySpace.NormalizeName(name)
	if e != nil || !creationKey.MatchString(key) {
		return repositoryDomainEntitySpace.Space{}, repositoryDomainEntitySpace.ErrInput
	}
	return s.Store.Create(ctx, r, actor, key, name)
}
