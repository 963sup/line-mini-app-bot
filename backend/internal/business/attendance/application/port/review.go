package port

import (
	"time"

	attendanceentity "github.com/963sup/line-mini-app-bot/internal/business/attendance/domain/entity"
	attendancepolicy "github.com/963sup/line-mini-app-bot/internal/business/attendance/domain/policy"
)

// ReviewTransaction supplies one consistent view and commits both writes atomically.
type ReviewTransaction interface {
	ReviewerGroups() ([]string, error)
	PreviousDecision() (attendancepolicy.Decision, bool, error)
	Original(attendanceentity.Correction) (attendanceentity.Record, error)
	SaveApproval(attendanceentity.Correction, attendanceentity.Record, attendanceentity.Record, string, time.Time) error
	SaveDecision(attendancepolicy.Decision, string, time.Time) error
}
