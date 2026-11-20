package server

import (
	"context"
	"errors"

	groupdb "github.com/963sup/line-mini-app-bot/internal/business/group/adapter/firestore"
	groupquery "github.com/963sup/line-mini-app-bot/internal/business/group/application/query"
	"github.com/963sup/line-mini-app-bot/internal/business/project"
	projectdb "github.com/963sup/line-mini-app-bot/internal/business/project/firestore"
	"github.com/963sup/line-mini-app-bot/internal/business/repository"
)

type projectSources struct{ repositories *repository.Service }

func (s projectSources) Authorize(ctx context.Context, actor, id, expectedOwnerID string) (project.Source, error) {
	if s.repositories == nil {
		return project.Source{}, project.ErrUnavailable
	}
	r, e := s.repositories.Get(ctx, actor, id)
	if errors.Is(e, repository.ErrNotFound) {
		return project.Source{}, project.ErrNotFound
	}
	if e != nil {
		return project.Source{}, e
	}
	if expectedOwnerID != "" && r.OwnerID != expectedOwnerID {
		return project.Source{}, project.ErrNotFound
	}
	return project.Source{ID: r.ID, OwnerID: r.OwnerID}, nil
}

func newProjects(groups *groupquery.Service, repositories *repository.Service) *project.Service {
	if groups == nil {
		return nil
	}
	store, ok := groups.Store.(*groupdb.Store)
	if !ok {
		return nil
	}
	return &project.Service{Groups: groups, Sources: projectSources{repositories}, Store: &projectdb.Store{Client: store.Client, ChannelID: store.ChannelID}}
}
