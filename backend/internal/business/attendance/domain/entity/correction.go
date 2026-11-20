package entity

import (
	"errors"
	"time"

	attendancevalueobject "github.com/963sup/line-mini-app-bot/internal/business/attendance/domain/valueobject"
)

var ErrCorrection = errors.New("invalid, stale or unauthorized attendance correction")

var ErrIdempotencyConflict = errors.New("idempotency key already used for another correction")

type Correction struct {
	ApplicantName         string
	NewRecord             bool
	ID                    string
	UserKey               string
	GroupID               string
	RecordID              string
	ExpectedEventSequence int64
	ExpectedRevision      int64
	Reason                string
	Start                 time.Time
	End                   time.Time
	Breaks                []attendancevalueobject.Interval
	RestConfirmed         bool
	CreatedAt             time.Time
}
