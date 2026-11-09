// Package attendance defines the group-scoped attendance contract.
package attendance

import (
	"context"
	"errors"
	"time"

	"github.com/963sup/line-mini-app-bot/internal/worktime"
)

var ErrInvalidRequest = errors.New("invalid or expired attendance request")

type Position struct {
	Latitude       float64   `json:"latitude" firestore:"latitude"`
	Longitude      float64   `json:"longitude" firestore:"longitude"`
	AccuracyMeters float64   `json:"accuracyMeters" firestore:"accuracyMeters"`
	CapturedAt     time.Time `json:"capturedAt" firestore:"capturedAt"`
}

type Request struct {
	GroupID   string    `firestore:"groupID"`
	UserID    string    `firestore:"userKey"`
	EventID   string    `firestore:"eventID"`
	Action    string    `firestore:"action"`
	EventTime int64     `firestore:"eventTime"`
	ExpiresAt time.Time `firestore:"expiresAt"`
}

// LocatedStore binds a browser submission to its original signed group event.
type LocatedStore interface {
	Submit(context.Context, string, string, Position, time.Time) (Result, error)
	Groups(context.Context, string) ([]string, error)
	RecentGroup(context.Context, string, string) ([]Record, error)
}

type LaunchStore interface {
	CreateLaunch(context.Context, string, string) (string, error)
	LaunchGroup(context.Context, string) (string, error)
	ExchangeLaunch(context.Context, string, string) (string, error)
}

type Record struct {
	Revision         int64               `firestore:"revision"`
	CorrectionID     string              `firestore:"correctionID,omitempty"`
	RebuildIssues    []string            `firestore:"-"`
	LastAppliedAt    time.Time           `firestore:"lastAppliedAt"`
	SourceEventIDs   []string            `firestore:"sourceEventIDs,omitempty"`
	EventSequence    int64               `firestore:"eventSequence"`
	RecordID         string              `firestore:"recordID,omitempty"`
	Breaks           []worktime.Interval `firestore:"breaks,omitempty"`
	BreakStartedAt   time.Time           `firestore:"breakStartedAt,omitempty"`
	RestConfirmed    bool                `firestore:"restConfirmed"`
	SchemaVersion    int                 `firestore:"schemaVersion"`
	GroupID          string              `firestore:"groupID"`
	ClockInAt        time.Time           `firestore:"clockInAt"`
	ClockOutAt       time.Time           `firestore:"clockOutAt,omitempty"`
	ClockInPosition  *Position           `firestore:"clockInPosition,omitempty"`
	ClockOutPosition *Position           `firestore:"clockOutPosition,omitempty"`
}

type Result struct {
	Code   string `firestore:"code"`
	Record Record `firestore:"record"`
}

type Store interface {
	Punch(context.Context, string, string, string, string, time.Time, int64) (Result, error)
	Recent(context.Context, string) ([]Record, error)
}
