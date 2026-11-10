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
	GroupBoard(context.Context, string, time.Time, time.Time) (Board, error)
	CurrentShift(context.Context, string, string) (*Record, error)
}
