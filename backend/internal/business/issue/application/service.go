package application

import (
	"github.com/963sup/line-mini-app-bot/internal/business/issue/application/command"
	"github.com/963sup/line-mini-app-bot/internal/business/issue/application/port"
	"github.com/963sup/line-mini-app-bot/internal/business/issue/application/query"
)

type (
	Commands = command.Operations
	Queries  = query.Operations
	Service  struct {
		Commands
		Queries
	}
)

func New(repos port.Repositories, store port.Store) *Service {
	a := port.Access{Repositories: repos, Store: store}
	return &Service{Commands: command.Operations{Access: a}, Queries: query.Operations{Access: a}}
}
