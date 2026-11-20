package statistics

import (
	"context"
	repositoryDomainEntity "github.com/963sup/line-mini-app-bot/internal/business/repository/domain/entity"
	repositoryDomainEntityStatistics "github.com/963sup/line-mini-app-bot/internal/business/repository/domain/entity/statistics"
)

type Repositories interface {
	Get(context.Context, string, string) (repositoryDomainEntity.Repository, error)
}

type Store interface {
	Statistics(context.Context, repositoryDomainEntity.Repository) (repositoryDomainEntityStatistics.Statistics, error)
}
