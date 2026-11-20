package dto

import (
	"errors"

	attendanceentity "github.com/963sup/line-mini-app-bot/internal/business/attendance/domain/entity"
)

var ErrInvalidCursor = errors.New("invalid attendance cursor")

// GroupPage contains the indexed query result and its continuation cursor.
type GroupPage struct {
	Items      []string
	NextCursor string
	QueryMode  string
}

type RecordPage struct {
	Items      []attendanceentity.Record
	NextCursor string
	QueryMode  string
}
