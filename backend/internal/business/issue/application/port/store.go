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
		Update(context.Context, repositoryDomainEntity.Repository, string, string, entity.Update) (entity.Issue, error)
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
