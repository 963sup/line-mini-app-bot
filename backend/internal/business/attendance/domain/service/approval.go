package service

import (
	attendanceentity "github.com/963sup/line-mini-app-bot/internal/business/attendance/domain/entity"
)

// VerifyCorrection checks the latest approval against its before/after evidence.
// It does not require a new ledger or traverse historical approvals.
func VerifyCorrection(before attendanceentity.Record, c attendanceentity.Correction, approved, current attendanceentity.Record) bool {
	expected, err := ApplyCorrection(before, c)
	if err != nil {
		return false
	}
	same := func(a, b attendanceentity.Record) bool {
		if a.RecordID != b.RecordID || a.GroupID != b.GroupID || a.Revision != b.Revision || a.CorrectionID != b.CorrectionID || a.EventSequence != b.EventSequence || !a.ClockInAt.Equal(b.ClockInAt) || !a.ClockOutAt.Equal(b.ClockOutAt) || a.RestConfirmed != b.RestConfirmed || !a.BreakStartedAt.Equal(b.BreakStartedAt) || len(a.Breaks) != len(b.Breaks) || len(a.SourceEventIDs) != len(b.SourceEventIDs) {
			return false
		}
		for i, p := range a.Breaks {
			if !p.Start.Equal(b.Breaks[i].Start) || !p.End.Equal(b.Breaks[i].End) {
				return false
			}
		}
		for i, id := range a.SourceEventIDs {
			if id != b.SourceEventIDs[i] {
				return false
			}
		}
		return true
	}
	return same(expected, approved) && same(approved, current)
}
