package attendance

import (
	"errors"
	"github.com/963sup/line-mini-app-bot/internal/worktime"
	"strings"
	"time"
)

var ErrCorrection = errors.New("invalid, stale or unauthorized attendance correction")

type Correction struct {
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

// ApplyCorrection creates a new projection; original events remain unchanged.
func ApplyCorrection(before Record, c Correction) (Record, error) {
	if c.ID == "" || c.RecordID != before.RecordID || c.GroupID != before.GroupID || c.ExpectedRevision != before.Revision || c.ExpectedEventSequence != before.EventSequence || strings.TrimSpace(c.Reason) == "" || len(c.Reason) > 2000 || !c.RestConfirmed {
		return Record{}, ErrCorrection
	}
	session := worktime.Session{ID: c.RecordID, Start: c.Start, End: c.End, Breaks: c.Breaks, RestConfirmed: c.RestConfirmed}
	r := worktime.Calculate([]worktime.Session{session}, c.Start, c.End)
	if len(r.Issues) > 0 || c.Start.IsZero() || !c.End.After(c.Start) {
		return Record{}, ErrCorrection
	}
	after := before
	after.ClockInAt, after.ClockOutAt = c.Start, c.End
	after.Breaks = append([]worktime.Interval(nil), c.Breaks...)
	after.BreakStartedAt = time.Time{}
	after.RestConfirmed = true
	after.Revision++
	after.CorrectionID = c.ID
	after.RebuildIssues = nil
	return after, nil
}

// VerifyCorrection checks the latest approval against its before/after evidence.
// It does not require a new ledger or traverse historical approvals.
func VerifyCorrection(before Record, c Correction, approved, current Record) bool {
	expected, err := ApplyCorrection(before, c)
	if err != nil {
		return false
	}
	same := func(a, b Record) bool {
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
