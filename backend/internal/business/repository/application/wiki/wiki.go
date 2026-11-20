package wiki

import (
	repositoryApplication "github.com/963sup/line-mini-app-bot/internal/business/repository/application"
	repositoryApplicationPortWiki "github.com/963sup/line-mini-app-bot/internal/business/repository/application/port/wiki"
)

type Service struct {
	Repositories *repositoryApplication.Service
	Store        repositoryApplicationPortWiki.Store
}
