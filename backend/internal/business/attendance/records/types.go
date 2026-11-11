package records

import (
	"time"

	"github.com/963sup/line-mini-app-bot/internal/business/attendance/worktime"
)

type Position struct {
	Latitude       float64
	Longitude      float64
	AccuracyMeters float64
	CapturedAt     time.Time
}

type Record struct {
	IsOpen           bool
	Revision         int64
	CorrectionID     string
	RebuildIssues    []string
	LastAppliedAt    time.Time
	SourceEventIDs   []string
	EventSequence    int64
	RecordID         string
	Breaks           []worktime.Interval
	BreakStartedAt   time.Time
	RestConfirmed    bool
	SchemaVersion    int
	GroupID          string
	ClockInAt        time.Time
	ClockOutAt       time.Time
	ClockInPosition  *Position
	ClockOutPosition *Position
}
