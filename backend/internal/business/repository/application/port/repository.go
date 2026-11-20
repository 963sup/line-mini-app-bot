package port

import (
	"context"

	repositoryApplicationDto "github.com/963sup/line-mini-app-bot/internal/business/repository/application/dto"
	repositoryDomainEntity "github.com/963sup/line-mini-app-bot/internal/business/repository/domain/entity"
)

type Store interface {
	Observe(context.Context, repositoryDomainEntity.Observation) error
	GetByID(context.Context, string) (repositoryDomainEntity.Repository, error)
	GetByGroupID(context.Context, string) (repositoryDomainEntity.Repository, error)
	Candidates(context.Context, string, string) (repositoryApplicationDto.Page, error)
}

type Membership interface {
	IsMember(context.Context, string, string) (bool, error)
}
