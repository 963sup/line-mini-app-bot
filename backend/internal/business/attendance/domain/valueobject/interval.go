package valueobject

import (
	"time"
)

type Interval struct{ Start, End time.Time }

// ValidWithin checks a completed rest interval without changing the value.
// Incomplete intervals remain representable as evidence and fail this check.
func (v Interval) ValidWithin(start, end time.Time) bool {
	return !v.Start.Before(start) && !v.End.After(end) && v.End.After(v.Start)
}
