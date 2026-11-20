package bootstrap

import (
	groupdb "github.com/963sup/line-mini-app-bot/internal/business/group/adapter/firestore"
	groupquery "github.com/963sup/line-mini-app-bot/internal/business/group/application/query"
	projectdb "github.com/963sup/line-mini-app-bot/internal/business/project/adapter/firestore"
	source "github.com/963sup/line-mini-app-bot/internal/business/project/adapter/repository"
	projectApplication "github.com/963sup/line-mini-app-bot/internal/business/project/application"
	repositoryApplication "github.com/963sup/line-mini-app-bot/internal/business/repository/application"
)

func newProjects(groups *groupquery.Service, store *groupdb.Store, repositories *repositoryApplication.Service) *projectApplication.Service {
	if store == nil {
		return nil
	}
	return &projectApplication.Service{Groups: groups, Sources: source.Sources{Repositories: repositories}, Store: &projectdb.Store{Client: store.Client, ChannelID: store.ChannelID}}
}
