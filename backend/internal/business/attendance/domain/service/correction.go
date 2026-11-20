package service

import (
	"strings"
	"time"

	attendanceentity "github.com/963sup/line-mini-app-bot/internal/business/attendance/domain/entity"
	attendancevalueobject "github.com/963sup/line-mini-app-bot/internal/business/attendance/domain/valueobject"
)

// ApplyCorrection creates a new projection; original events remain unchanged.
func ApplyCorrection(before attendanceentity.Record, c attendanceentity.Correction) (attendanceentity.Record, error) {
	if c.NewRecord {
		if before.RecordID != "" || !before.ClockInAt.IsZero() || before.Revision != 0 || c.ExpectedRevision != 0 || c.ExpectedEventSequence != 0 {
			return attendanceentity.Record{}, attendanceentity.ErrCorrection
		}
		before = attendanceentity.Record{RecordID: c.RecordID, GroupID: c.GroupID, SchemaVersion: 3}
	}
	if c.RecordID == "" || c.GroupID == "" || c.ExpectedRevision < 0 || c.ExpectedEventSequence < 0 || len(c.Breaks) > 100 || c.ID == "" || c.RecordID != before.RecordID || c.GroupID != before.GroupID || c.ExpectedRevision != before.Revision || c.ExpectedEventSequence != before.EventSequence || strings.TrimSpace(c.Reason) == "" || len(c.Reason) > 2000 || !c.RestConfirmed {
		return attendanceentity.Record{}, attendanceentity.ErrCorrection
	}
	session := attendanceentity.Session{ID: c.RecordID, Start: c.Start, End: c.End, Breaks: c.Breaks, RestConfirmed: c.RestConfirmed}
	r := Calculate([]attendanceentity.Session{session}, c.Start, c.End)
	if len(r.Issues) > 0 || c.Start.IsZero() || !c.End.After(c.Start) {
		return attendanceentity.Record{}, attendanceentity.ErrCorrection
	}
	after := before
	after.ClockInAt, after.ClockOutAt = c.Start, c.End
	after.IsOpen = false
	after.Breaks = append([]attendancevalueobject.Interval(nil), c.Breaks...)
	after.BreakStartedAt = time.Time{}
	after.RestConfirmed = true
	after.Revision++
	after.CorrectionID = c.ID
	after.LastAppliedAt = c.End
	after.RebuildIssues = nil
	return after, nil
}
