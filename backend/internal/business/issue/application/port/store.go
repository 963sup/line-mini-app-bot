package port

import (
	repositoryDomainEntity "github.com/963sup/line-mini-app-bot/internal/business/repository/domain/entity"
)

import (
	"context"
	"errors"

	"github.com/963sup/line-mini-app-bot/internal/business/issue/application/dto"
	"github.com/963sup/line-mini-app-bot/internal/business/issue/domain/entity"
)

var ErrUnavailable = errors.New("issue storage unavailable")

type (
	Repositories interface {
		Get(context.Context, string, string) (repositoryDomainEntity.Repository, error)
	}
	Store interface {
		List(context.Context, repositoryDomainEntity.Repository, string, dto.Filter) (dto.Page, error)
		Get(context.Context, repositoryDomainEntity.Repository, string) (entity.Issue, error)
		Create(context.Context, repositoryDomainEntity.Repository, string, string, entity.Content) (entity.Issue, error)
		RunUpdate(ctx context.Context, scope repositoryDomainEntity.Repository, id string, update func(UpdateWork) (entity.Issue, error)) (entity.Issue, error)
	}
)

type Access struct {
	Repositories Repositories
	Store        Store
}

func (s *Access) Scope(ctx context.Context, actor, repositoryID string) (repositoryDomainEntity.Repository, error) {
	if s == nil || s.Store == nil || s.Repositories == nil {
		return repositoryDomainEntity.Repository{}, ErrUnavailable
	}
	r, err := s.Repositories.Get(ctx, actor, repositoryID)
	if errors.Is(err, repositoryDomainEntity.ErrNotFound) {
		return repositoryDomainEntity.Repository{}, entity.ErrNotFound
	}
	return r, err
}

type UpdateWork interface {
	CheckRepository(actor string) error
	Current() (entity.Issue, error)
	ValidateLabels([]string) error
	ValidateMilestone(string) error
	Save(actor string, previous, updated entity.Issue) error
}
