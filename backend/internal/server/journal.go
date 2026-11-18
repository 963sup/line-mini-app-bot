package server

import (
	groupdomain "github.com/963sup/line-mini-app-bot/internal/business/group"
	groupdb "github.com/963sup/line-mini-app-bot/internal/business/group/firestore"
	"github.com/963sup/line-mini-app-bot/internal/business/journal"
	journaldb "github.com/963sup/line-mini-app-bot/internal/business/journal/firestore"
	"github.com/963sup/line-mini-app-bot/internal/business/project"
	"github.com/963sup/line-mini-app-bot/internal/business/repository"
)

func newJournals(groups *groupdomain.Service, repositories *repository.Service, projects *project.Service) journal.Operations {
	if groups == nil || repositories == nil || projects == nil {
		return nil
	}
	store, ok := groups.Store.(*groupdb.Store)
	if !ok {
		return nil
	}
	return &journaldb.Store{Client: store.Client, Channel: store.ChannelID, Groups: groups, Repositories: repositories, Projects: projects}
}
