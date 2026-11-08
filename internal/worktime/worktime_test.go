package worktime

import (
	"testing"
	"time"
)

func at(s string) time.Time {
	t, err := time.Parse(time.RFC3339, s)
	if err != nil {
		panic(err)
	}
	return t
}
func TestCalculateBreaksOverlapAndMonth(t *testing.T) {
	from := at("2026-11-01T00:00:00+08:00")
	to := from.AddDate(0, 1, 0)
	sessions := []Session{{ID: "overnight", Start: at("2026-10-31T22:00:00+08:00"), End: at("2026-11-01T06:00:00+08:00"), RestConfirmed: true, Breaks: []Interval{{at("2026-11-01T01:00:00+08:00"), at("2026-11-01T02:00:00+08:00")}, {at("2026-11-01T01:30:00+08:00"), at("2026-11-01T02:30:00+08:00")}}}}
	result := Calculate(sessions, from, to)
	if result.Minutes != 270 || result.Pending != 0 {
		t.Fatalf("%+v", result)
	}
	previous := Calculate(sessions, from.AddDate(0, -1, 0), from)
	if previous.Minutes != 120 {
		t.Fatalf("previous month: %+v", previous)
	}
	// Overlapping work intervals are counted once and reported.
	sessions = append(sessions, Session{ID: "duplicate", Start: at("2026-11-01T03:00:00+08:00"), End: at("2026-11-01T04:00:00+08:00"), RestConfirmed: true})
	result = Calculate(sessions, from, to)
	if result.Minutes != 270 || len(result.Issues) != 1 {
		t.Fatalf("overlap: %+v", result)
	}
}
func TestIncompleteAndLegacyAreNotGuessed(t *testing.T) {
	start := at("2026-11-01T09:00:00+08:00")
	end := start.Add(8 * time.Hour)
	for _, s := range []Session{{Start: start, RestConfirmed: true}, {Start: start, End: end}, {Start: start, End: start.Add(-time.Minute), RestConfirmed: true}, {Start: start, End: end, RestConfirmed: true, Breaks: []Interval{{Start: start.Add(time.Hour)}}}} {
		r := Calculate([]Session{s}, start.Add(-24*time.Hour), end.Add(24*time.Hour))
		if r.Minutes != 0 || r.Pending != 1 || len(r.Issues) != 1 {
			t.Fatalf("%+v", r)
		}
	}
}
func TestBuildMultipleShiftsAndDuplicates(t *testing.T) {
	start := at("2026-11-01T09:00:00+08:00")
	events := []Event{{ID: "in", Kind: "clock-in", At: start}, {ID: "out", Kind: "clock-out", At: start.Add(4 * time.Hour)}, {ID: "confirm", Kind: "confirm-rest", At: start.Add(3 * time.Hour)}, {ID: "in2", Kind: "clock-in", At: start.Add(5 * time.Hour)}, {ID: "rest", Kind: "break-start", At: start.Add(6 * time.Hour)}, {ID: "rest-end", Kind: "break-end", At: start.Add(7 * time.Hour)}, {ID: "confirm2", Kind: "confirm-rest", At: start.Add(8 * time.Hour)}, {ID: "out2", Kind: "clock-out", At: start.Add(9 * time.Hour)}}
	events = append(events, events[0])
	sessions, issues := BuildIntervals(events)
	r := Calculate(sessions, start, start.Add(24*time.Hour))
	if len(issues) != 0 || len(sessions) != 2 || r.Minutes != 420 {
		t.Fatalf("sessions=%+v issues=%+v result=%+v", sessions, issues, r)
	}
	events = append(events, Event{ID: "in", Kind: "clock-in", At: start.Add(time.Minute)})
	sessions, issues = BuildIntervals(events)
	r = Calculate(sessions, start, start.Add(24*time.Hour))
	if len(issues) == 0 || r.Minutes != 0 {
		t.Fatal("conflicting event was silently counted")
	}
}
func TestAccumulateSecondsBeforeConvertingMinutes(t *testing.T) {
	start := at("2026-11-01T09:00:00+08:00")
	r := Calculate([]Session{{Start: start, End: start.Add(40 * time.Second), RestConfirmed: true}, {Start: start.Add(time.Minute), End: start.Add(100 * time.Second), RestConfirmed: true}}, start, start.Add(time.Hour))
	if r.Duration != 80*time.Second || r.Minutes != 1 {
		t.Fatalf("%+v", r)
	}
}

func TestApplyAndReplayAgreeAfterRestChanges(t *testing.T) {
	start := at("2026-11-01T09:00:00+08:00")
	events := []Event{
		{ID: "in", Kind: "clock-in", At: start},
		{ID: "confirm", Kind: "confirm-rest", At: start.Add(time.Hour)},
		{ID: "rest", Kind: "break-start", At: start.Add(2 * time.Hour)},
		{ID: "rest-end", Kind: "break-end", At: start.Add(3 * time.Hour)},
		{ID: "out", Kind: "clock-out", At: start.Add(8 * time.Hour)},
	}
	state := State{}
	var live Session
	for i := range events {
		events[i].Sequence = int64(i + 1)
		next, out := ApplyEvent(state, events[i])
		if !out.Accepted {
			t.Fatal(out)
		}
		events[i].ResultCode = out.Code
		state = next
		live = out.Session
	}
	sessions, issues := BuildIntervals(events)
	if len(issues) != 0 || len(sessions) != 1 || live.RestConfirmed || sessions[0].RestConfirmed {
		t.Fatalf("live=%+v replay=%+v issues=%v", live, sessions, issues)
	}
	if len(live.EventIDs) != 5 || len(sessions[0].EventIDs) != 5 {
		t.Fatal("missing source evidence")
	}
	if len(live.Breaks) != 1 || !live.Breaks[0].End.Equal(sessions[0].Breaks[0].End) {
		t.Fatal("rest mismatch")
	}
}
func TestRejectedEventsDoNotPoisonReplay(t *testing.T) {
	start := at("2026-11-01T09:00:00+08:00")
	events := []Event{{ID: "in", Kind: "clock-in", At: start, ResultCode: "clocked_in", Sequence: 1}, {ID: "reject", Kind: "clock-out", At: start, ResultCode: "stale_event", Sequence: 1}, {ID: "confirm", Kind: "confirm-rest", At: start.Add(time.Hour), ResultCode: "rest_confirmed", Sequence: 2}, {ID: "out", Kind: "clock-out", At: start.Add(8 * time.Hour), ResultCode: "clocked_out", Sequence: 3}}
	sessions, issues := BuildIntervals(events)
	r := Calculate(sessions, start, start.Add(24*time.Hour))
	if len(issues) != 0 || r.Minutes != 480 || len(r.EventIDs) != 3 {
		t.Fatalf("%+v %+v", issues, r)
	}
}
