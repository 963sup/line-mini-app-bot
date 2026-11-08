package attendance

import (
	"github.com/963sup/line-mini-app-bot/internal/worktime"
	"time"
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
		breaks := append([]worktime.Interval(nil), r.Breaks...)
		if !r.BreakStartedAt.IsZero() {
			breaks = append(breaks, worktime.Interval{Start: r.BreakStartedAt})
		}
		sessions = append(sessions, worktime.Session{ID: r.RecordID, Start: r.ClockInAt, End: r.ClockOutAt, Breaks: breaks, RestConfirmed: r.RestConfirmed, EventIDs: r.SourceEventIDs})
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
