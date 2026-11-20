package space

import (
	"context"

	repositoryApplicationDtoSpace "github.com/963sup/line-mini-app-bot/internal/business/repository/application/dto/space"
	repositoryDomainEntity "github.com/963sup/line-mini-app-bot/internal/business/repository/domain/entity"
	repositoryDomainEntitySpace "github.com/963sup/line-mini-app-bot/internal/business/repository/domain/entity/space"
)

type Repositories interface {
	Get(context.Context, string, string) (repositoryDomainEntity.Repository, error)
}

type Store interface {
	List(context.Context, repositoryDomainEntity.Repository, string, string) (repositoryApplicationDtoSpace.Page, error)
	Get(context.Context, repositoryDomainEntity.Repository, string) (repositoryDomainEntitySpace.Space, error)
	Create(context.Context, repositoryDomainEntity.Repository, string, string, string) (repositoryDomainEntitySpace.Space, error)
}
