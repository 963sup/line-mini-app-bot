package port

import (
	"context"

	objectDomainEntity "github.com/963sup/line-mini-app-bot/internal/business/object/domain/entity"
	objectimportDomainEntity "github.com/963sup/line-mini-app-bot/internal/business/objectimport/domain/entity"
	"github.com/963sup/line-mini-app-bot/internal/business/repository"
)

type Repositories interface {
	Get(context.Context, string, string) (repository.Repository, error)
}

type Objects interface {
	Create(context.Context, string, string, string, objectDomainEntity.Content) (objectDomainEntity.Object, error)
}

type Store interface {
	Create(context.Context, repository.Repository, string, string, string, objectimportDomainEntity.Job) (objectimportDomainEntity.Job, error)
	Get(context.Context, repository.Repository, string, string) (objectimportDomainEntity.Job, error)
	Review(context.Context, repository.Repository, string, string, objectimportDomainEntity.Review) (objectimportDomainEntity.Job, error)
	BeginConfirm(context.Context, repository.Repository, string, string, int64) (objectimportDomainEntity.Job, error)
	RecordResult(context.Context, repository.Repository, string, string, string, string, int64) (objectimportDomainEntity.Job, error)
}
