package port

import (
	"context"

	projectDomainEntity "github.com/963sup/line-mini-app-bot/internal/business/project/domain/entity"
)

// ReferenceStateStore reads identity only, without accessing source content.
type ReferenceStateStore interface {
	ReferencePresent(context.Context, Scope, string, string, projectDomainEntity.Reference) (bool, error)
}
