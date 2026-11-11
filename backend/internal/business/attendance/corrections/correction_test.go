package corrections

import (
	"testing"
	"time"

	recordmodel "github.com/963sup/line-mini-app-bot/internal/business/attendance/records"
	"github.com/963sup/line-mini-app-bot/internal/business/attendance/worktime"
)

func TestCorrectionIsVersionedAndDoesNotMutateOriginal(t *testing.T) {
	start := time.Date(2026, 11, 1, 9, 0, 0, 0, time.FixedZone("Taipei", 28800))
	before := recordmodel.Record{RecordID: "r", GroupID: "g", ClockInAt: start, SourceEventIDs: []string{"event"}}
	c := Correction{ID: "c", GroupID: "g", RecordID: "r", Reason: "漏打下班", Start: start, End: start.Add(8 * time.Hour), RestConfirmed: true}
	after, err := ApplyCorrection(before, c)
	if err != nil || after.Revision != 1 || after.CorrectionID != "c" || !before.ClockOutAt.IsZero() {
		t.Fatalf("%+v %v", after, err)
	}
	if _, err := ApplyCorrection(after, c); err == nil {
		t.Fatal("stale correction accepted")
	}
	report := recordmodel.CalculateMonthlyAttendance([]recordmodel.Record{after}, time.Date(2026, 11, 1, 0, 0, 0, 0, start.Location()))
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
