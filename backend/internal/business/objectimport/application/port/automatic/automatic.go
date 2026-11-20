package automatic

import (
	repositoryDomainEntity "github.com/963sup/line-mini-app-bot/internal/business/repository/domain/entity"
)

import (
	"context"

	objectimportDomainEntity "github.com/963sup/line-mini-app-bot/internal/business/objectimport/domain/entity"
	objectimportDomainEntityAutomatic "github.com/963sup/line-mini-app-bot/internal/business/objectimport/domain/entity/automatic"
)

type Producer interface {
	Rows(context.Context, objectimportDomainEntity.Source) ([]objectimportDomainEntityAutomatic.Candidate, error)
}

// Update must load/check scope and ownership, then atomically persist the callback result.
// A nil current run represents absence; only the initial intent operation accepts absence.
type Store interface {
	Update(context.Context, repositoryDomainEntity.Repository, string, string, func(*objectimportDomainEntityAutomatic.Run) (objectimportDomainEntityAutomatic.Run, error)) (objectimportDomainEntityAutomatic.Run, error)
}
