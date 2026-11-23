package entity

import (
	"errors"
	"time"
)

var (
	ErrInput    = errors.New("invalid journal")
	ErrNotFound = errors.New("journal or source unavailable to actor")
	ErrConflict = errors.New("idempotency key already used for different content")
)

type (
	Reference struct{ Kind, ID string }
	Work      struct {
		Name       string
		Percent    int
		Reference  *Reference
		ObjectID   string
		SpaceIDs   []string
		ObjectName string
		SpaceNames []string
	}
)

type (
	Personnel struct{ Management, Control, Workers []string }
	Draft     struct {
		GroupID         string
		RepositoryID    string
		WorkDate        time.Time
		Personnel       Personnel
		Today, Tomorrow []Work
	}
)

// ApplyReferenceCheck preserves the journaled work when a reader loses source access.
func (w *Work) ApplyReferenceCheck(err error, redactMissing bool) error {
	if !redactMissing || !errors.Is(err, ErrNotFound) {
		return err
	}
	w.Reference = nil
	w.ObjectID = ""
	w.SpaceIDs = nil
	w.ObjectName = ""
	w.SpaceNames = nil
	return nil
}

type Entry struct {
	Draft
	ID, CreatorID string
	CreatedAt     time.Time
}
