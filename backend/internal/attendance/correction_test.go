package attendance

import (
	"testing"
	"time"

	"github.com/963sup/line-mini-app-bot/internal/attendance/worktime"
)

func TestCorrectionIsVersionedAndDoesNotMutateOriginal(t *testing.T) {
	start := time.Date(2026, 11, 1, 9, 0, 0, 0, time.FixedZone("Taipei", 28800))
	before := Record{RecordID: "r", GroupID: "g", ClockInAt: start, SourceEventIDs: []string{"event"}}
	c := Correction{ID: "c", GroupID: "g", RecordID: "r", Reason: "漏打下班", Start: start, End: start.Add(8 * time.Hour), RestConfirmed: true}
	after, err := ApplyCorrection(before, c)
	if err != nil || after.Revision != 1 || after.CorrectionID != "c" || !before.ClockOutAt.IsZero() {
		t.Fatalf("%+v %v", after, err)
	}
	if _, err := ApplyCorrection(after, c); err == nil {
		t.Fatal("stale correction accepted")
	}
	report := CalculateMonthlyAttendance([]Record{after}, time.Date(2026, 11, 1, 0, 0, 0, 0, start.Location()))
	if report.Calculation.ConfirmedNetMinutes != 480 || len(report.CorrectionIDs) != 1 {
		t.Fatalf("%+v", report)
	}
	changed := before
	changed.EventSequence = 1
	if _, err := ApplyCorrection(changed, c); err == nil {
		t.Fatal("proposal overwrote newer punch events")
	}
	c.Breaks = []worktime.Interval{{Start: start.Add(-time.Hour), End: start}}
	if _, err := ApplyCorrection(before, c); err == nil {
		t.Fatal("out-of-shift rest accepted")
	}
}

func TestVerifyCorrectionRejectsChangedEvidence(t *testing.T) {
	start := time.Date(2026, 11, 1, 9, 0, 0, 0, time.UTC)
	before := Record{RecordID: "r", GroupID: "g", ClockInAt: start}
	c := Correction{ID: "c", RecordID: "r", GroupID: "g", Reason: "補下班", Start: start, End: start.Add(8 * time.Hour), RestConfirmed: true}
	after, err := ApplyCorrection(before, c)
	if err != nil {
		t.Fatal(err)
	}
	if !VerifyCorrection(before, c, after, after) {
		t.Fatal("valid approval rejected")
	}
	changed := after
	changed.ClockOutAt = changed.ClockOutAt.Add(time.Hour)
	if VerifyCorrection(before, c, after, changed) {
		t.Fatal("changed projection accepted")
	}
	c.Reason = ""
	if VerifyCorrection(before, c, after, after) {
		t.Fatal("invalid proposal accepted")
	}
}
