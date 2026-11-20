package entity

import (
	"time"
)

// State is the single transition contract for both live writes and replay.
type State struct {
	Active        *Session
	RestStartedAt time.Time
	LastAt        time.Time
}
