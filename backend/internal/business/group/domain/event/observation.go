package event

import (
	"time"
)

// Observation is accepted only from the signature-verified webhook adapter.
// Users are discovery candidates, never an authorization grant.
type Observation struct {
	GroupID, Name, Status, EventID string
	At                             time.Time
	Users                          []string
}
