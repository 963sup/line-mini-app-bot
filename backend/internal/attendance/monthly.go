package attendance

import (
	"time"

	"github.com/963sup/line-mini-app-bot/internal/attendance/worktime"
)

type MonthlyReport struct {
	Start         time.Time
	Records       []Record
	Calculation   worktime.Result
	CorrectionIDs []string
}

// CalculateMonthlyAttendance is shared by chat, details and exports.
func CalculateMonthlyAttendance(records []Record, start time.Time) MonthlyReport {
	report := MonthlyReport{Start: start, Records: append([]Record(nil), records...)}
	sessions := make([]worktime.Session, 0, len(records))
	for _, r := range records {
		sessions = append(sessions, recordSession(r))
		if r.CorrectionID != "" {
			report.CorrectionIDs = append(report.CorrectionIDs, r.CorrectionID)
		}
	}
	report.Calculation = worktime.Calculate(sessions, start, start.AddDate(0, 1, 0))
	for _, r := range records {
		for _, code := range r.RebuildIssues {
			report.Calculation.Issues = append(report.Calculation.Issues, worktime.Issue{SessionID: r.RecordID, Code: code})
		}
	}
	return report
}

func recordSession(r Record) worktime.Session {
	breaks := append([]worktime.Interval(nil), r.Breaks...)
	if !r.BreakStartedAt.IsZero() {
		breaks = append(breaks, worktime.Interval{Start: r.BreakStartedAt})
	}
	return worktime.Session{
		ID: r.RecordID, Start: r.ClockInAt, End: r.ClockOutAt, Breaks: breaks, RestConfirmed: r.RestConfirmed,
		EventIDs: r.SourceEventIDs, Invalid: len(r.RebuildIssues) > 0,
	}
}

// CalculateRecordAttendance uses the same rules as monthly totals over the full shift.
func CalculateRecordAttendance(r Record) worktime.Result {
	end := r.ClockOutAt
	if !end.After(r.ClockInAt) {
		// Keep a nonempty calculation window so invalid clock times are reported.
		end = r.ClockInAt.Add(time.Nanosecond)
	}
	return worktime.Calculate([]worktime.Session{recordSession(r)}, r.ClockInAt, end)
}
