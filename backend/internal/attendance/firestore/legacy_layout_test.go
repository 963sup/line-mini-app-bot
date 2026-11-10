// Frozen pre-boundary Firestore/JSON format; intentionally independent of production codecs.
package firestore

import (
	"time"

	"github.com/963sup/line-mini-app-bot/internal/attendance/worktime"
)

type legacyPosition struct {
	Latitude       float64   `json:"latitude" firestore:"latitude"`
	Longitude      float64   `json:"longitude" firestore:"longitude"`
	AccuracyMeters float64   `json:"accuracyMeters" firestore:"accuracyMeters"`
	CapturedAt     time.Time `json:"capturedAt" firestore:"capturedAt"`
}
type legacyRecord struct {
	IsOpen           bool                `firestore:"isOpen" json:"-"`
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
	ClockInPosition  *legacyPosition     `firestore:"clockInPosition,omitempty"`
	ClockOutPosition *legacyPosition     `firestore:"clockOutPosition,omitempty"`
}
type legacyRequest struct {
	GroupID   string    `firestore:"groupID"`
	UserID    string    `firestore:"userKey"`
	EventID   string    `firestore:"eventID"`
	Action    string    `firestore:"action"`
	EventTime int64     `firestore:"eventTime"`
	ExpiresAt time.Time `firestore:"expiresAt"`
}
type legacyResult struct {
	Code   string       `firestore:"code"`
	Record legacyRecord `firestore:"record"`
}
type legacyCorrection struct {
	ID                    string              `firestore:"id" json:"id"`
	UserKey               string              `firestore:"userKey" json:"-"`
	GroupID               string              `firestore:"groupID" json:"groupId"`
	RecordID              string              `firestore:"recordID" json:"recordId"`
	ExpectedEventSequence int64               `firestore:"expectedEventSequence" json:"expectedEventSequence"`
	ExpectedRevision      int64               `firestore:"expectedRevision" json:"expectedRevision"`
	Reason                string              `firestore:"reason" json:"reason"`
	Start                 time.Time           `firestore:"start" json:"start"`
	End                   time.Time           `firestore:"end" json:"end"`
	Breaks                []worktime.Interval `firestore:"breaks" json:"breaks"`
	RestConfirmed         bool                `firestore:"restConfirmed" json:"restConfirmed"`
	CreatedAt             time.Time           `firestore:"createdAt" json:"-"`
}
