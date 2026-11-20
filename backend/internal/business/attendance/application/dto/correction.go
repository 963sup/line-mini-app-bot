package dto

import (
	"time"

	attendanceentity "github.com/963sup/line-mini-app-bot/internal/business/attendance/domain/entity"
)

// CorrectionDetails describes a request and its human decision, never an effective shift.
type CorrectionDetails struct {
	attendanceentity.Correction
	Status         string
	DecisionReason string
	ReviewedAt     time.Time
	Before         *attendanceentity.Record
}

type CorrectionPage struct {
	Items      []CorrectionDetails
	NextCursor string
}
