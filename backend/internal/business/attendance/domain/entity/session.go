package entity

import (
	"time"

	attendancevalueobject "github.com/963sup/line-mini-app-bot/internal/business/attendance/domain/valueobject"
)

type Session struct {
	ID            string
	Start, End    time.Time
	Breaks        []attendancevalueobject.Interval
	RestConfirmed bool
	Invalid       bool
	EventIDs      []string
}
