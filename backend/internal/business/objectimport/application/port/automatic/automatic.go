package automatic

import (
	"context"

	objectimportDomainEntity "github.com/963sup/line-mini-app-bot/internal/business/objectimport/domain/entity"
	objectimportDomainEntityAutomatic "github.com/963sup/line-mini-app-bot/internal/business/objectimport/domain/entity/automatic"
	"github.com/963sup/line-mini-app-bot/internal/business/repository"
)

type Producer interface {
	Rows(context.Context, objectimportDomainEntity.Source) ([]objectimportDomainEntityAutomatic.Candidate, error)
}

// Update must load/check scope and ownership, then atomically persist the callback result.
// A nil current run represents absence; only the initial intent operation accepts absence.
type Store interface {
	Update(context.Context, repository.Repository, string, string, func(*objectimportDomainEntityAutomatic.Run) (objectimportDomainEntityAutomatic.Run, error)) (objectimportDomainEntityAutomatic.Run, error)
}
