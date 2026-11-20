package port

import (
	"context"
	"errors"
	"github.com/963sup/line-mini-app-bot/internal/business/issue/application/dto"
	"github.com/963sup/line-mini-app-bot/internal/business/issue/domain/entity"
	"github.com/963sup/line-mini-app-bot/internal/business/repository"
)

var ErrUnavailable = errors.New("issue storage unavailable")

type (
	Repositories interface {
		Get(context.Context, string, string) (repository.Repository, error)
	}
	Store interface {
		List(context.Context, repository.Repository, string, dto.Filter) (dto.Page, error)
		Get(context.Context, repository.Repository, string) (entity.Issue, error)
		Create(context.Context, repository.Repository, string, string, entity.Content) (entity.Issue, error)
		Update(context.Context, repository.Repository, string, string, entity.Update) (entity.Issue, error)
	}
)

type Access struct {
	Repositories Repositories
	Store        Store
}

func (s *Access) Scope(ctx context.Context, actor, repositoryID string) (repository.Repository, error) {
	if s == nil || s.Store == nil || s.Repositories == nil {
		return repository.Repository{}, ErrUnavailable
	}
	r, err := s.Repositories.Get(ctx, actor, repositoryID)
	if errors.Is(err, repository.ErrNotFound) {
		return repository.Repository{}, entity.ErrNotFound
	}
	return r, err
}
