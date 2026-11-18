// Package journal defines the operations shared by Feed and Dashboard.
package journal

import (
	"context"
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
		Name      string
		Percent   int
		Reference *Reference
	}
)

type (
	Personnel struct{ Management, Control, Workers []string }
	Draft     struct {
		SpaceID         string
		WorkDate        time.Time
		Personnel       Personnel
		Today, Tomorrow []Work
	}
)

type Entry struct {
	Draft
	ID, CreatorID string
	CreatedAt     time.Time
}
type (
	Filter struct{ SpaceID, WorkDate, Cursor string }
	Page   struct {
		Items      []Entry
		NextCursor string
	}
)

// Operations is implemented by the membership-aware persistent service.
// Actor is always supplied by the authenticated transport. Implementations
// authorize each read, redact inaccessible references and scope create retries
// by actor and key before returning data.
type Operations interface {
	Create(ctx context.Context, actor, key string, draft Draft) (Entry, error)
	Get(ctx context.Context, actor, id string) (Entry, error)
	List(ctx context.Context, actor string, filter Filter) (Page, error)
}
