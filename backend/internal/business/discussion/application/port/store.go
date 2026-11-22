package port

import (
	repositoryDomainEntity "github.com/963sup/line-mini-app-bot/internal/business/repository/domain/entity"
)

import (
	"context"
	"errors"

	"github.com/963sup/line-mini-app-bot/internal/business/discussion/application/dto"
	"github.com/963sup/line-mini-app-bot/internal/business/discussion/domain/entity"
)

var ErrUnavailable = errors.New("discussion storage unavailable")

type (
	Repositories interface {
		Get(context.Context, string, string) (repositoryDomainEntity.Repository, error)
	}
	Store interface {
		List(context.Context, repositoryDomainEntity.Repository, string, dto.Filter) (dto.Page, error)
		Get(context.Context, repositoryDomainEntity.Repository, string) (entity.Discussion, error)
		Create(context.Context, repositoryDomainEntity.Repository, string, string, entity.Content) (entity.Discussion, error)
		RunUpdate(ctx context.Context, scope repositoryDomainEntity.Repository, id string, update func(UpdateWork) (entity.Discussion, error)) (entity.Discussion, error)
		ListComments(context.Context, repositoryDomainEntity.Repository, string, string, string) (dto.CommentPage, error)
		CreateComment(context.Context, repositoryDomainEntity.Repository, string, string, string, string) (entity.Comment, error)
		UpdateComment(context.Context, repositoryDomainEntity.Repository, string, string, string, string, int64) (entity.Comment, error)
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
	Repository() (id, owner, status string, err error)
	Current() (entity.Discussion, error)
	ValidateLabels([]string) error
	Save(actor string, previous, updated entity.Discussion) error
}
