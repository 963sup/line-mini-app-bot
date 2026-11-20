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

const retention = 24 * time.Hour

func (m Message) CanCollect(now time.Time) bool {
	return m.ID != "" && m.UserID != "" && !m.OccurredAt.Before(now.Add(-retention)) && !m.OccurredAt.After(now.Add(time.Minute))
}

// Retain assigns the collection epoch and keeps expiry relative to the original event.
func (m *Message) Retain(epoch int64) {
	m.Epoch = epoch
	m.ExpireAt = m.OccurredAt.Add(retention)
}

// ApplyEdit returns true when the retained message accepts this edit.
func (m *Message) ApplyEdit(text string, timestamp, epoch int64, now time.Time) bool {
	if m.Deleted || m.Epoch != epoch || !m.ExpireAt.After(now) || timestamp <= m.EditedAt || timestamp < m.OccurredAt.UnixMilli() {
		return false
	}
	m.Text, m.EditedAt = text, timestamp
	return true
}

// Store retains messages and invalidates summaries when a source is withdrawn.
// Editing remains an optional capability of the storage adapter.
type Store interface {
	Collect(context.Context, string, Message) error
	Unsend(context.Context, string, string) error
}
