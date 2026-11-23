package port

import (
	repositoryDomainEntity "github.com/963sup/line-mini-app-bot/internal/business/repository/domain/entity"
)

import (
	"context"

	objectApplicationDto "github.com/963sup/line-mini-app-bot/internal/business/object/application/dto"
	objectDomainEntity "github.com/963sup/line-mini-app-bot/internal/business/object/domain/entity"
	"github.com/963sup/line-mini-app-bot/internal/business/object/domain/event"
)

type Repositories interface {
	Get(context.Context, string, string) (repositoryDomainEntity.Repository, error)
}

type Store interface {
	List(context.Context, repositoryDomainEntity.Repository, string, string, ...string) (objectApplicationDto.Page, error)
	Get(context.Context, repositoryDomainEntity.Repository, string) (objectDomainEntity.Object, error)
	Create(context.Context, repositoryDomainEntity.Repository, string, string, objectDomainEntity.Content) (objectDomainEntity.Object, error)
	Update(context.Context, repositoryDomainEntity.Repository, string, string, objectDomainEntity.Update) (objectDomainEntity.Object, error)
}

type QuantityHistory interface {
	ListQuantityChanges(context.Context, repositoryDomainEntity.Repository, string, string) (objectApplicationDto.QuantityChangePage, error)
}

// QuantityEvents reads and acknowledges committed revisions in the configured channel.
type QuantityEvents interface {
	PendingQuantityEvents(context.Context) ([]event.PendingRevision, error)
	AcknowledgeQuantityEvent(context.Context, string) error
}
