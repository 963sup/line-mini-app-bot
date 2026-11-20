package server

import (
	projectApplication "github.com/963sup/line-mini-app-bot/internal/business/project/application"
	projectDomainEntity "github.com/963sup/line-mini-app-bot/internal/business/project/domain/entity"
)

import (
	"context"
	"errors"

	groupdb "github.com/963sup/line-mini-app-bot/internal/business/group/adapter/firestore"
	groupquery "github.com/963sup/line-mini-app-bot/internal/business/group/application/query"

	projectdb "github.com/963sup/line-mini-app-bot/internal/business/project/adapter/firestore"
	"github.com/963sup/line-mini-app-bot/internal/business/repository"
)

type projectSources struct{ repositories *repository.Service }

func (s projectSources) Authorize(ctx context.Context, actor, id, expectedOwnerID string) (projectDomainEntity.Source, error) {
	if s.repositories == nil {
		return projectDomainEntity.Source{}, projectDomainEntity.ErrUnavailable
	}
	r, e := s.repositories.Get(ctx, actor, id)
	if errors.Is(e, repository.ErrNotFound) {
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

func newProjects(groups *groupquery.Service, repositories *repository.Service) *projectApplication.Service {
	if groups == nil {
		return nil
	}
	store, ok := groups.Store.(*groupdb.Store)
	if !ok {
		return nil
	}
	return &projectApplication.Service{Groups: groups, Sources: projectSources{repositories}, Store: &projectdb.Store{Client: store.Client, ChannelID: store.ChannelID}}
}
