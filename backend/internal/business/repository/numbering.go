package repository

import (
	"errors"
	"regexp"
	"time"
)

// MaxResourceNumber is the largest integer represented exactly by browser clients.
const MaxResourceNumber int64 = 1<<53 - 1

var (
	ErrNumberedCreation = errors.New("invalid numbered resource creation")
	ErrCreationConflict = errors.New("creation key already used for different content")
	ErrNumberExhausted  = errors.New("repository resource numbers exhausted")
	resourceTypePattern = regexp.MustCompile(`^[a-z][a-z0-9-]{0,63}$`)
	creationKeyPattern  = regexp.MustCompile(`^[A-Za-z0-9_-]{16,128}$`)
	payloadHashPattern  = regexp.MustCompile(`^[a-f0-9]{64}$`)
)

// NumberedCreation is an internal adapter command, never a client-owned identity
// or an authorization grant. The caller must authorize the specific business
// operation first and hash its normalized, versioned business input.
type NumberedCreation struct {
	RepositoryID, GroupID, ActorID string
	ResourceType, Key, PayloadHash string
}

func (c NumberedCreation) Valid() bool {
	return ValidID(c.RepositoryID) && ValidGroup(c.GroupID) && c.ActorID != "" && len(c.ActorID) <= 256 &&
		resourceTypePattern.MatchString(c.ResourceType) && creationKeyPattern.MatchString(c.Key) && payloadHashPattern.MatchString(c.PayloadHash)
}

// NumberedResource identifies a resource independently of its domain state. All
// resource types in one repository share Number; IDs remain opaque and stable.
type NumberedResource struct {
	RepositoryID, ResourceType, ResourceID string
	Number                                 int64
	CreatedAt                              time.Time
}

func (r NumberedResource) Valid() bool {
	return ValidID(r.RepositoryID) && ValidID(r.ResourceID) && resourceTypePattern.MatchString(r.ResourceType) &&
		r.Number > 0 && r.Number <= MaxResourceNumber && !r.CreatedAt.IsZero()
}
