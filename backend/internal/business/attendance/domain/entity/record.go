package entity

import (
	"time"

	attendancevalueobject "github.com/963sup/line-mini-app-bot/internal/business/attendance/domain/valueobject"
)

type Record struct {
	IsOpen           bool
	Revision         int64
	CorrectionID     string
	RebuildIssues    []string
	LastAppliedAt    time.Time
	SourceEventIDs   []string
	EventSequence    int64
	RecordID         string
	Breaks           []attendancevalueobject.Interval
	BreakStartedAt   time.Time
	RestConfirmed    bool
	SchemaVersion    int
	GroupID          string
	ClockInAt        time.Time
	ClockOutAt       time.Time
	ClockInPosition  *attendancevalueobject.Position
	ClockOutPosition *attendancevalueobject.Position
}
