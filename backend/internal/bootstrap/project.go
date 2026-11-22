package bootstrap

import (
	projectdb "github.com/963sup/line-mini-app-bot/internal/business/project/adapter/firestore"
	projectgroup "github.com/963sup/line-mini-app-bot/internal/business/project/adapter/group"
	source "github.com/963sup/line-mini-app-bot/internal/business/project/adapter/repository"
	projectApplication "github.com/963sup/line-mini-app-bot/internal/business/project/application"
	repositoryApplication "github.com/963sup/line-mini-app-bot/internal/business/repository/application"
	groupdb "github.com/963sup/line-mini-app-bot/internal/linegroup/adapter/firestore"
	groupquery "github.com/963sup/line-mini-app-bot/internal/linegroup/application/query"
)

func newProjects(groups *groupquery.Service, store *groupdb.Store, repositories *repositoryApplication.Service) *projectApplication.Service {
	if store == nil {
		return nil
	}
	return &projectApplication.Service{Groups: projectgroup.Access{Groups: groups}, Sources: source.Sources{Repositories: repositories}, Store: &projectdb.Store{Client: store.Client, ChannelID: store.ChannelID}}
}
