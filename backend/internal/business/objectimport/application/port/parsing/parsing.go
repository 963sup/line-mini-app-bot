package parsing

import (
	repositoryDomainEntity "github.com/963sup/line-mini-app-bot/internal/business/repository/domain/entity"
)

import (
	"context"

	objectimportDomainEntity "github.com/963sup/line-mini-app-bot/internal/business/objectimport/domain/entity"
)

type Producer interface {
	Candidates(context.Context, objectimportDomainEntity.Source) ([]objectimportDomainEntity.Candidate, error)
}

type Repositories interface {
	Get(context.Context, string, string) (repositoryDomainEntity.Repository, error)
}
