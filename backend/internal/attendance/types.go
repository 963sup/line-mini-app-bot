// Package attendance defines the group-scoped attendance contract.
package attendance

import (
	"errors"
	"time"

	"github.com/963sup/line-mini-app-bot/internal/attendance/worktime"
)

var ErrExpiredLaunch = errors.New("launch expired")

var ErrInvalidRequest = errors.New("invalid or expired attendance request")

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

type Result struct {
	Code   string
	Record Record
}
