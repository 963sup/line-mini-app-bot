package board

import (
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
