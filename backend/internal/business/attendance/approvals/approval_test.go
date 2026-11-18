package approvals

import (
	"testing"
	"time"

	correctionsmodel "github.com/963sup/line-mini-app-bot/internal/business/attendance/corrections"
	recordmodel "github.com/963sup/line-mini-app-bot/internal/business/attendance/records"
)

func TestVerifyCorrectionRejectsChangedEvidence(t *testing.T) {
	start := time.Date(2026, 11, 1, 9, 0, 0, 0, time.UTC)
	before := recordmodel.Record{RecordID: "r", GroupID: "g", ClockInAt: start}
	c := correctionsmodel.Correction{ID: "c", RecordID: "r", GroupID: "g", Reason: "補下班", Start: start, End: start.Add(8 * time.Hour), RestConfirmed: true}
	after, err := correctionsmodel.ApplyCorrection(before, c)
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
