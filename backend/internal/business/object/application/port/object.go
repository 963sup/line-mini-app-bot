package port

import (
	"context"

	objectApplicationDto "github.com/963sup/line-mini-app-bot/internal/business/object/application/dto"
	objectDomainEntity "github.com/963sup/line-mini-app-bot/internal/business/object/domain/entity"
	"github.com/963sup/line-mini-app-bot/internal/business/repository"
)

type Repositories interface {
	Get(context.Context, string, string) (repository.Repository, error)
}

type Store interface {
	List(context.Context, repository.Repository, string, string) (objectApplicationDto.Page, error)
	Get(context.Context, repository.Repository, string) (objectDomainEntity.Object, error)
	Create(context.Context, repository.Repository, string, string, objectDomainEntity.Content) (objectDomainEntity.Object, error)
	Update(context.Context, repository.Repository, string, string, objectDomainEntity.Update) (objectDomainEntity.Object, error)
}
