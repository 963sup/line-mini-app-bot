package attendance

import (
	"context"
	"errors"
	"time"
)

var ErrBoardLimit = errors.New("attendance board read limit exceeded")

// Board counts distinct people in disjoint current states, not shifts or expected staff.
type Board struct {
	Working  int
	Resting  int
	Finished int
	AsOf     time.Time
}

type BoardStore interface {
	GroupBoard(ctx context.Context, groupID string, start, now time.Time) (Board, error)
	CurrentShift(ctx context.Context, groupID, userID string) (*Record, error)
}
