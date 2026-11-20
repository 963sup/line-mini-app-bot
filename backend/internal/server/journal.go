package server

import (
	projectApplication "github.com/963sup/line-mini-app-bot/internal/business/project/application"
)

import (
	groupdb "github.com/963sup/line-mini-app-bot/internal/business/group/adapter/firestore"
	groupquery "github.com/963sup/line-mini-app-bot/internal/business/group/application/query"

	journaldb "github.com/963sup/line-mini-app-bot/internal/business/journal/adapter/firestore"
	journalport "github.com/963sup/line-mini-app-bot/internal/business/journal/application/port"

	"github.com/963sup/line-mini-app-bot/internal/business/repository"
)

func newJournals(groups *groupquery.Service, repositories *repository.Service, projects *projectApplication.Service) journalport.Operations {
	if groups == nil || repositories == nil || projects == nil {
		return nil
	}
	store, ok := groups.Store.(*groupdb.Store)
	if !ok {
		return nil
	}
	return &journaldb.Store{Client: store.Client, Channel: store.ChannelID, Groups: groups, Repositories: repositories, Projects: projects}
}
