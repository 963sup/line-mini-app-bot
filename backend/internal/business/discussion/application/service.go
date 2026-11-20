package application

import (
	"github.com/963sup/line-mini-app-bot/internal/business/discussion/application/command"
	"github.com/963sup/line-mini-app-bot/internal/business/discussion/application/port"
	"github.com/963sup/line-mini-app-bot/internal/business/discussion/application/query"
)

type (
	Commands = command.Service
	Queries  = query.Service
	Service  struct {
		Commands
		Queries
	}
)

func New(repositories port.Repositories, store port.Store) *Service {
	access := port.Access{Repositories: repositories, Store: store}
	return &Service{command.Service{Access: access}, query.Service{Access: access}}
}
