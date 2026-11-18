package messagecollection

import (
	"context"
	"time"
)

type Message struct {
	ID, EventID, UserID, Text string
	OccurredAt, ExpireAt      time.Time
	Epoch                     int64
	Deleted                   bool
	EditedAt                  int64
}

// Store retains messages and invalidates summaries when a source is withdrawn.
// Editing remains an optional capability of the storage adapter.
type Store interface {
	Collect(context.Context, string, Message) error
	Unsend(context.Context, string, string) error
}
