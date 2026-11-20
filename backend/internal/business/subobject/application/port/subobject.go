package port

import (
	"context"
	repositoryDomainEntity "github.com/963sup/line-mini-app-bot/internal/business/repository/domain/entity"
	subobjectApplicationDto "github.com/963sup/line-mini-app-bot/internal/business/subobject/application/dto"
	subobjectDomainEntity "github.com/963sup/line-mini-app-bot/internal/business/subobject/domain/entity"
)

type Repositories interface {
	Get(context.Context, string, string) (repositoryDomainEntity.Repository, error)
}

type Store interface {
	List(context.Context, repositoryDomainEntity.Repository, string, string, string) (subobjectApplicationDto.Page, error)
	Get(context.Context, repositoryDomainEntity.Repository, string, string, string) (subobjectDomainEntity.Subobject, error)
	Create(context.Context, repositoryDomainEntity.Repository, string, string, string, subobjectDomainEntity.Content) (subobjectDomainEntity.Subobject, error)
	Update(context.Context, repositoryDomainEntity.Repository, string, string, string, subobjectDomainEntity.Update) (subobjectDomainEntity.Subobject, error)
}
