package port

import (
	repositoryDomainEntity "github.com/963sup/line-mini-app-bot/internal/business/repository/domain/entity"
)

import (
	"context"

	objectDomainEntity "github.com/963sup/line-mini-app-bot/internal/business/object/domain/entity"
	objectimportDomainEntity "github.com/963sup/line-mini-app-bot/internal/business/objectimport/domain/entity"
)

type Repositories interface {
	Get(context.Context, string, string) (repositoryDomainEntity.Repository, error)
}

type Objects interface {
	Create(context.Context, string, string, string, objectDomainEntity.Content) (objectDomainEntity.Object, error)
}

type Store interface {
	Create(context.Context, repositoryDomainEntity.Repository, string, string, string, objectimportDomainEntity.Job) (objectimportDomainEntity.Job, error)
	Get(context.Context, repositoryDomainEntity.Repository, string, string) (objectimportDomainEntity.Job, error)
	Review(context.Context, repositoryDomainEntity.Repository, string, string, objectimportDomainEntity.Review) (objectimportDomainEntity.Job, error)
	BeginConfirm(context.Context, repositoryDomainEntity.Repository, string, string, int64) (objectimportDomainEntity.Job, error)
	RecordResult(context.Context, repositoryDomainEntity.Repository, string, string, string, string, int64) (objectimportDomainEntity.Job, error)
}
