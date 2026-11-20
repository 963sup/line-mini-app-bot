package wiki

import (
	"context"
	repositoryDomainEntity "github.com/963sup/line-mini-app-bot/internal/business/repository/domain/entity"
	repositoryDomainEntityWiki "github.com/963sup/line-mini-app-bot/internal/business/repository/domain/entity/wiki"
)

type Store interface {
	GetWiki(context.Context, repositoryDomainEntity.Repository, string) (repositoryDomainEntityWiki.Page, error)
	ListWiki(context.Context, repositoryDomainEntity.Repository, string, string, string) ([]repositoryDomainEntityWiki.Page, string, error)
	SaveWiki(context.Context, repositoryDomainEntity.Repository, string, string, string, int64, string, string) (repositoryDomainEntityWiki.Page, error)
}
