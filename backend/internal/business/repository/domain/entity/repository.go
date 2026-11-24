package entity

import (
	"errors"
	"regexp"
	"time"

	"github.com/963sup/line-mini-app-bot/internal/platform/line/lineid"
)

var ErrNotFound = errors.New("repository not found")

var ErrUnavailable = errors.New("repository unavailable")

var ErrCursor = errors.New("invalid repository cursor")

var idPattern = regexp.MustCompile(`^[0-9a-f]{32}$`)

func ValidOwner(id string) bool { return lineid.ValidUser(id) || lineid.ValidGroup(id) }

func ValidID(id string) bool { return idPattern.MatchString(id) }

type Repository struct {
	ID, OwnerID, Name, Status string
	Version                   int64
	CreatedAt, UpdatedAt      time.Time
}

// MatchesActive reports whether the current stored repository still matches this scope.
func (r Repository) MatchesActive(id, ownerID, status string) bool {
	return r.ID == id && r.OwnerID == ownerID && status == "ACTIVE"
}

// Observation is accepted only from the signature-verified webhook adapter.
// Users are discovery candidates, never an authorization grant.
type Observation struct {
	OwnerID, Name, Status, EventID string
	At                             time.Time
	Users                          []string
}
