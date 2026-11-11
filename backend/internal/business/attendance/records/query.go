package records

import "errors"

var (
	ErrInvalidCursor  = errors.New("invalid attendance cursor")
	ErrRecordNotFound = errors.New("attendance record not found")
)

// QueryMode exposes the temporary historical-read fallback until migration completes.
type GroupPage struct {
	Items      []string
	NextCursor string
	QueryMode  string
}

type RecordPage struct {
	Items      []Record
	NextCursor string
	QueryMode  string
}
