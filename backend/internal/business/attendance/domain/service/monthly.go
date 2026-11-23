package service

import (
	"time"

	attendanceentity "github.com/963sup/line-mini-app-bot/internal/business/attendance/domain/entity"
	attendancevalueobject "github.com/963sup/line-mini-app-bot/internal/business/attendance/domain/valueobject"
)

type MonthlyReport struct {
	Start         time.Time
	Records       []attendanceentity.Record
	Calculation   Result
	CorrectionIDs []string
}

// CalculateMonthlyAttendance is shared by chat, details and exports.
func CalculateMonthlyAttendance(records []attendanceentity.Record, start time.Time) MonthlyReport {
	report := MonthlyReport{Start: start, Records: append([]attendanceentity.Record(nil), records...)}
	sessions := make([]attendanceentity.Session, 0, len(records))
	for _, r := range records {
		sessions = append(sessions, recordSession(r))
		if r.CorrectionID != "" {
			report.CorrectionIDs = append(report.CorrectionIDs, r.CorrectionID)
		}
	}
	report.Calculation = Calculate(sessions, start, start.AddDate(0, 1, 0))
	for _, r := range records {
		for _, code := range r.RebuildIssues {
			report.Calculation.Issues = append(report.Calculation.Issues, Issue{SessionID: r.RecordID, Code: code})
		}
	}
	return report
}

func recordSession(r attendanceentity.Record) attendanceentity.Session {
	breaks := append([]attendancevalueobject.Interval(nil), r.Breaks...)
	if !r.BreakStartedAt.IsZero() {
		breaks = append(breaks, attendancevalueobject.Interval{Start: r.BreakStartedAt})
	}
	return attendanceentity.Session{
		ID: r.RecordID, Start: r.ClockInAt, End: r.ClockOutAt, Breaks: breaks, RestConfirmed: r.RestConfirmed,
		EventIDs: r.SourceEventIDs, Invalid: len(r.RebuildIssues) > 0,
	}
}

const (
	RecordInvalid   = "invalid"
	RecordEstimated = "estimated"
	RecordConfirmed = "confirmed"
)

// CalculateRecordAttendance uses the same rules as monthly totals over the full shift.
func CalculateRecordAttendance(r attendanceentity.Record) (Result, string) {
	end := r.ClockOutAt
	if !end.After(r.ClockInAt) {
		// Keep a nonempty calculation window so invalid clock times are reported.
		end = r.ClockInAt.Add(time.Nanosecond)
	}
	work := Calculate([]attendanceentity.Session{recordSession(r)}, r.ClockInAt, end)
	if work.Uncalculable > 0 {
		return work, RecordInvalid
	}
	if work.Unconfirmed > 0 {
		return work, RecordEstimated
	}
	return work, RecordConfirmed
}
