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
	Candidates(context.Context, string, string, string) (repositoryApplicationDto.Page, error)
}

type AddressWriter interface {
	UpdateAddress(context.Context, repositoryDomainEntity.Repository, repositoryDomainEntity.Repository) (repositoryDomainEntity.Repository, error)
}

type OwnerAccess interface {
	CanAccessOwner(context.Context, string, string, string) (bool, error)
}
