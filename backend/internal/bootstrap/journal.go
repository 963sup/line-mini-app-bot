package bootstrap

import (
	groupdb "github.com/963sup/line-mini-app-bot/internal/business/group/adapter/firestore"
	groupquery "github.com/963sup/line-mini-app-bot/internal/business/group/application/query"
	journaldb "github.com/963sup/line-mini-app-bot/internal/business/journal/adapter/firestore"
	journalport "github.com/963sup/line-mini-app-bot/internal/business/journal/application/port"
	projectApplication "github.com/963sup/line-mini-app-bot/internal/business/project/application"
	repositoryApplication "github.com/963sup/line-mini-app-bot/internal/business/repository/application"
)

func newJournals(groups *groupquery.Service, store *groupdb.Store, repositories *repositoryApplication.Service, projects *projectApplication.Service) journalport.Operations {
	if store == nil || repositories == nil || projects == nil {
		return nil
	}
	return &journaldb.Store{Client: store.Client, Channel: store.ChannelID, Groups: groups, Repositories: repositories, Projects: projects}
}
