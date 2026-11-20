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
		Name        string
		Percent     int
		Reference   *Reference
		ObjectID    string   `json:",omitempty"`
		SubobjectID string   `json:",omitempty"`
		SpaceIDs    []string `json:",omitempty"`
		ObjectName  string   `json:",omitempty"`
		SpaceNames  []string `json:",omitempty"`
	}
)

type (
	Personnel struct{ Management, Control, Workers []string }
	Draft     struct {
		GroupID         string
		WorkDate        time.Time
		Personnel       Personnel
		Today, Tomorrow []Work
	}
)

// CheckReferences checks both days. On reads, redactMissing hides all details
// derived from an inaccessible source while preserving the journaled work.
func (d Draft) CheckReferences(redactMissing bool, check func(Reference) error) error {
	for _, rows := range [][]Work{d.Today, d.Tomorrow} {
		for i := range rows {
			w := &rows[i]
			if w.Reference == nil {
				continue
			}
			err := check(*w.Reference)
			if redactMissing && errors.Is(err, ErrNotFound) {
				w.Reference = nil
				w.ObjectID = ""
				w.SubobjectID = ""
				w.SpaceIDs = nil
				w.ObjectName = ""
				w.SpaceNames = nil
				continue
			}
			if err != nil {
				return err
			}
		}
	}
	return nil
}

type Entry struct {
	Draft
	ID, CreatorID string
	CreatedAt     time.Time
}
