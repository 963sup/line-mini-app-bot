package wiki

import (
	"context"

	repositoryDomainEntity "github.com/963sup/line-mini-app-bot/internal/business/repository/domain/entity"
	repositoryDomainEntityWiki "github.com/963sup/line-mini-app-bot/internal/business/repository/domain/entity/wiki"
)

type WriteTarget struct {
	ID, Actor, CreationKey string
}

type WriteState struct {
	ID                     string
	Current, FirstRevision *repositoryDomainEntityWiki.Page
}

type DecideWrite func(WriteState) (page repositoryDomainEntityWiki.Page, changed bool, err error)

type Repositories interface {
	Get(context.Context, string, string) (repositoryDomainEntity.Repository, error)
}

type Store interface {
	GetWiki(context.Context, repositoryDomainEntity.Repository, string) (repositoryDomainEntityWiki.Page, error)
	ListWiki(context.Context, repositoryDomainEntity.Repository, string, string, string) ([]repositoryDomainEntityWiki.Page, string, error)
	// WriteWiki decides against one transactional snapshot. It may retry the
	// decision and must commit the page, title reservation, and revision together.
	WriteWiki(context.Context, repositoryDomainEntity.Repository, WriteTarget, DecideWrite) (repositoryDomainEntityWiki.Page, error)
}
