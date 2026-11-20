package statistics

import (
	"context"
	"errors"
	repositoryApplicationPortStatistics "github.com/963sup/line-mini-app-bot/internal/business/repository/application/port/statistics"
	repositoryDomainEntity "github.com/963sup/line-mini-app-bot/internal/business/repository/domain/entity"
	repositoryDomainEntityStatistics "github.com/963sup/line-mini-app-bot/internal/business/repository/domain/entity/statistics"
)

type Service struct {
	Repositories repositoryApplicationPortStatistics.Repositories
	Store        repositoryApplicationPortStatistics.Store
}

func (s *Service) Get(ctx context.Context, actor, id string) (repositoryDomainEntityStatistics.Statistics, error) {
	if s == nil || s.Repositories == nil || s.Store == nil {
		return repositoryDomainEntityStatistics.Statistics{}, repositoryDomainEntityStatistics.ErrUnavailable
	}
	r, e := s.Repositories.Get(ctx, actor, id)
	if errors.Is(e, repositoryDomainEntity.ErrNotFound) {
		return repositoryDomainEntityStatistics.Statistics{}, repositoryDomainEntityStatistics.ErrNotFound
	}
	if e != nil {
		return repositoryDomainEntityStatistics.Statistics{}, e
	}
	return s.Store.Statistics(ctx, r)
}
