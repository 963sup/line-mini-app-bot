package attendancehttp

import (
	"github.com/963sup/line-mini-app-bot/internal/attendance"
	"github.com/963sup/line-mini-app-bot/internal/attendance/worktime"
	"github.com/963sup/line-mini-app-bot/internal/httpapi/generated"
)

func intervalViews(source []worktime.Interval) []generated.Interval {
	if source == nil {
		return nil
	}
	result := make([]generated.Interval, len(source))
	for i, v := range source {
		result[i] = generated.Interval{Start: v.Start, End: v.End}
	}
	return result
}

func calculationView(source worktime.Result) generated.Calculation {
	intervals := intervalViews(source.Intervals)
	var issues []generated.CalculationIssue
	if source.Issues != nil {
		issues = make([]generated.CalculationIssue, len(source.Issues))
		for i, v := range source.Issues {
			issues[i] = generated.CalculationIssue{SessionID: v.SessionID, Code: v.Code}
		}
	}
	return generated.Calculation{AttendanceMinutes: &source.AttendanceMinutes, EstimatedMinutes: &source.EstimatedTotalMinutes, Unconfirmed: pointerInt64(source.Unconfirmed), Uncalculable: pointerInt64(source.Uncalculable), Duration: int64(source.Duration), EventIDs: &source.EventIDs, Intervals: &intervals, Issues: &issues, Minutes: source.ConfirmedNetMinutes, Pending: int64(source.Pending), Version: source.Version}
}

func optionalValue[T any](p *T) T {
	if p != nil {
		return *p
	}
	var zero T
	return zero
}

func correctionInput(source generated.CorrectionSubmission) attendance.Correction {
	var breaks []worktime.Interval
	if source.Breaks != nil && *source.Breaks != nil {
		breaks = make([]worktime.Interval, len(*source.Breaks))
		for i, v := range *source.Breaks {
			breaks[i] = worktime.Interval{Start: v.Start, End: v.End}
		}
	}
	return attendance.Correction{ID: optionalValue(source.Id), GroupID: source.GroupId, RecordID: source.RecordId, ExpectedEventSequence: optionalValue(source.ExpectedEventSequence), ExpectedRevision: optionalValue(source.ExpectedRevision), Reason: optionalValue(source.Reason), Start: optionalValue(source.Start), End: optionalValue(source.End), Breaks: breaks, RestConfirmed: optionalValue(source.RestConfirmed)}
}

func correctionView(source attendance.Correction) generated.CorrectionReview {
	breaks := intervalViews(source.Breaks)
	return generated.CorrectionReview{Id: source.ID, GroupId: source.GroupID, RecordId: source.RecordID, ExpectedEventSequence: source.ExpectedEventSequence, ExpectedRevision: source.ExpectedRevision, Reason: source.Reason, Start: source.Start, End: source.End, Breaks: &breaks, RestConfirmed: source.RestConfirmed}
}

func positionView(source *attendance.Position) *generated.NullablePosition {
	if source == nil {
		return nil
	}
	return &generated.NullablePosition{Latitude: source.Latitude, Longitude: source.Longitude, AccuracyMeters: source.AccuracyMeters, CapturedAt: source.CapturedAt}
}

func legacyRecordView(source attendance.Record) generated.StoredRecord {
	view := recordView(source)
	return generated.StoredRecord{BreakStartedAt: view.BreakStartedAt, Breaks: view.Breaks, ClockInAt: view.ClockInAt, ClockOutAt: view.ClockOutAt, EventSequence: view.EventSequence, GroupID: view.GroupID, RebuildIssues: view.RebuildIssues, RecordID: view.RecordID, RestConfirmed: view.RestConfirmed, Revision: view.Revision, CorrectionID: source.CorrectionID, LastAppliedAt: source.LastAppliedAt, SourceEventIDs: &source.SourceEventIDs, SchemaVersion: int64(source.SchemaVersion), ClockInPosition: positionView(source.ClockInPosition), ClockOutPosition: positionView(source.ClockOutPosition)}
}

func monthlyView(source attendance.MonthlyReport) generated.MonthlyReport {
	var records []generated.StoredRecord
	if source.Records != nil {
		records = make([]generated.StoredRecord, len(source.Records))
		for i, v := range source.Records {
			records[i] = legacyRecordView(v)
		}
	}
	return generated.MonthlyReport{Start: source.Start, Records: &records, Calculation: calculationView(source.Calculation), CorrectionIDs: &source.CorrectionIDs}
}

func pointerInt64(value int) *int64 { converted := int64(value); return &converted }
