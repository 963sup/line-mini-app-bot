package records

import "errors"

var (
	ErrInvalidCursor  = errors.New("invalid attendance cursor")
	ErrRecordNotFound = errors.New("attendance record not found")
)

// GroupPage contains the indexed query result and its continuation cursor.
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
