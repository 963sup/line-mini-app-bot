package entity

import (
	"errors"
	"regexp"
	"time"
)

var ErrNotFound = errors.New("repository not found")

var ErrUnavailable = errors.New("repository unavailable")

var ErrCursor = errors.New("invalid repository cursor")

var ErrInput = errors.New("invalid repository input")

var ErrConflict = errors.New("repository version conflict")

var idPattern = regexp.MustCompile(`^[0-9a-f]{32}$`)

// ownerPattern is the persisted owner-reference contract. Repository identity
// remains the independent ID validated by ValidID; an owner reference only
// scopes authorization and the external binding maintained by adapters.
var userOwnerPattern = regexp.MustCompile(`^U[0-9a-f]{32}$`)
var organizationOwnerPattern = regexp.MustCompile(`^[0-9a-f]{8}-[0-9a-f]{4}-[1-5][0-9a-f]{3}-[89ab][0-9a-f]{3}-[0-9a-f]{12}$`)

type OwnerType string

const (
	OwnerUser         OwnerType = "USER"
	OwnerOrganization OwnerType = "ORGANIZATION"
)

func ValidTypedOwner(ownerType OwnerType, id string) bool {
	switch ownerType {
	case OwnerUser:
		return userOwnerPattern.MatchString(id)
	case OwnerOrganization:
		return organizationOwnerPattern.MatchString(id)
	default:
		return false
	}
}

func ValidOwnerID(id string) bool {
	return ValidTypedOwner(OwnerUser, id) || ValidTypedOwner(OwnerOrganization, id)
}

func ValidOwner(id string) bool { return ValidOwnerID(id) }

func ValidID(id string) bool { return idPattern.MatchString(id) }

type Repository struct {
	ID, OwnerID, GroupID, Name, Status string
	OwnerType                          OwnerType
	Address                            *Address
	Version                            int64
	CreatedAt, UpdatedAt               time.Time
}

func (r Repository) ChangeAddress(expectedVersion int64, address Address, now time.Time) (Repository, error) {
	if r.Status != "ACTIVE" || r.Version != expectedVersion || now.IsZero() {
		return Repository{}, ErrConflict
	}
	if !address.Valid() {
		return Repository{}, ErrInput
	}
	r.Address, r.Version, r.UpdatedAt = &address, r.Version+1, now
	return r, nil
}

// MatchesActive reports whether the current stored repository still matches this scope.
func (r Repository) MatchesActive(id, ownerID, status string) bool {
	return r.ID == id && r.OwnerID == ownerID && status == "ACTIVE"
}

// Observation is accepted only from the signature-verified webhook adapter.
// Users are discovery candidates, never an authorization grant.
type Observation struct {
	OwnerID, GroupID, Name, Status, EventID string
	OwnerType                               OwnerType
	At                                      time.Time
	Users                                   []string
}
