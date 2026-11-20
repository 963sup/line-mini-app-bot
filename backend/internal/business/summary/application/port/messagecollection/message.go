package messagecollection

import (
	"context"

	summaryDomainEntityMessagecollection "github.com/963sup/line-mini-app-bot/internal/business/summary/domain/entity/messagecollection"
)

// Store retains messages and invalidates summaries when a source is withdrawn.
// Editing remains an optional capability of the storage adapter.
type Store interface {
	Collect(context.Context, string, summaryDomainEntityMessagecollection.Message) error
	Unsend(context.Context, string, string) error
}
