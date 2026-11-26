package port

import (
	"context"

	repositoryApplicationDto "github.com/963sup/line-mini-app-bot/internal/business/repository/application/dto"
	repositoryDomainEntity "github.com/963sup/line-mini-app-bot/internal/business/repository/domain/entity"
)

type Store interface {
	GetByID(context.Context, string) (repositoryDomainEntity.Repository, error)
	Candidates(context.Context, string, string, string) (repositoryApplicationDto.Page, error)
}

type BindingStore interface {
	ObservePersonal(context.Context, repositoryDomainEntity.PersonalObservation) error
	ObserveGroup(context.Context, repositoryDomainEntity.GroupObservation) error
	GetByGroupID(context.Context, string) (repositoryDomainEntity.Repository, error)
}

// CheckOrganization reads membership from the same snapshot as the repository.
type (
	CheckOrganization func(organizationID, actor string) (bool, error)
	AddressChange     func(repositoryDomainEntity.Repository, CheckOrganization) (repositoryDomainEntity.Repository, error)
)

type AddressWriter interface {
	RunAddressChange(context.Context, string, AddressChange) (repositoryDomainEntity.Repository, error)
}

type OwnerAccess interface {
	CanAccessOwner(context.Context, string, string, string) (bool, error)
}
