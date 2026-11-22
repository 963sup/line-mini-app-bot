package bootstrap

import (
	journaldb "github.com/963sup/line-mini-app-bot/internal/business/journal/adapter/firestore"
	journalport "github.com/963sup/line-mini-app-bot/internal/business/journal/application/port"
	journalquery "github.com/963sup/line-mini-app-bot/internal/business/journal/application/query"
	projectApplication "github.com/963sup/line-mini-app-bot/internal/business/project/application"
	repositoryApplication "github.com/963sup/line-mini-app-bot/internal/business/repository/application"
	groupdb "github.com/963sup/line-mini-app-bot/internal/linegroup/adapter/firestore"
	groupquery "github.com/963sup/line-mini-app-bot/internal/linegroup/application/query"
)

func newJournals(groups *groupquery.Service, store *groupdb.Store, repositories *repositoryApplication.Service, projects *projectApplication.Service) journalport.Operations {
	if store == nil || repositories == nil || projects == nil {
		return nil
	}
	return &journaldb.Store{Client: store.Client, Channel: store.ChannelID, ReferenceCatalog: journalquery.ReferenceCatalog{Groups: groups, Repositories: repositories, Projects: projects}}
}
