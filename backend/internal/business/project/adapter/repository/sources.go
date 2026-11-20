package repository

import (
	"context"
	"errors"

	projectDomainEntity "github.com/963sup/line-mini-app-bot/internal/business/project/domain/entity"
	repositoryApplication "github.com/963sup/line-mini-app-bot/internal/business/repository/application"
	repositoryDomainEntity "github.com/963sup/line-mini-app-bot/internal/business/repository/domain/entity"
)

type Sources struct {
	Repositories *repositoryApplication.Service
}

func (s Sources) Authorize(ctx context.Context, actor, id, expectedOwnerID string) (projectDomainEntity.Source, error) {
	if s.Repositories == nil {
		return projectDomainEntity.Source{}, projectDomainEntity.ErrUnavailable
	}
	r, e := s.Repositories.Get(ctx, actor, id)
	if errors.Is(e, repositoryDomainEntity.ErrNotFound) {
		return projectDomainEntity.Source{}, projectDomainEntity.ErrNotFound
	}
	if e != nil {
		return projectDomainEntity.Source{}, e
	}
	if expectedOwnerID != "" && r.OwnerID != expectedOwnerID {
		return projectDomainEntity.Source{}, projectDomainEntity.ErrNotFound
	}
	return projectDomainEntity.Source{ID: r.ID, OwnerID: r.OwnerID}, nil
}
