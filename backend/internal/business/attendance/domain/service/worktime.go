package service

import (
	"sort"
	"time"

	attendanceentity "github.com/963sup/line-mini-app-bot/internal/business/attendance/domain/entity"
	attendanceevent "github.com/963sup/line-mini-app-bot/internal/business/attendance/domain/event"
	attendancevalueobject "github.com/963sup/line-mini-app-bot/internal/business/attendance/domain/valueobject"
)

const Version = "worktime-v3"

type Issue struct{ SessionID, Code string }

type Result struct {
	EventIDs                                 []string
	Version                                  string
	Duration                                 time.Duration
	ConfirmedNetMinutes                      int64
	AttendanceMinutes, EstimatedTotalMinutes int64
	Unconfirmed, Uncalculable                int
	Pending                                  int
	Issues                                   []Issue
	Intervals                                []attendancevalueobject.Interval
}

// OtherNoticeCount excludes rest confirmation, which monthly reports display separately.
func (r Result) OtherNoticeCount() int {
	count := 0
	for _, issue := range r.Issues {
		if issue.Code != "rest_unconfirmed" {
			count++
		}
	}
	return count
}

// RebuildSessions preserves input, deduplicates event IDs and reports ambiguous sequences.
func RebuildSessions(events []attendanceevent.Event) ([]attendanceentity.Session, []Issue) {
	seen := map[string]attendanceevent.Event{}
	ordered := []attendanceevent.Event{}
	issues := []Issue{}
	conflict := false
	for _, e := range events {
		if e.ID == "" || e.At.IsZero() {
			issues = append(issues, Issue{Code: "invalid_event"})
			conflict = true
			continue
		}
		if old, ok := seen[e.ID]; ok {
			if old.Kind != e.Kind || !old.At.Equal(e.At) || old.Sequence != e.Sequence || old.ResultCode != e.ResultCode {
				issues = append(issues, Issue{Code: "conflicting_event"})
				conflict = true
			}
			continue
		}
		seen[e.ID] = e
		ordered = append(ordered, e)
	}
	sort.SliceStable(ordered, func(i, j int) bool {
		if ordered[i].Sequence != ordered[j].Sequence {
			return ordered[i].Sequence < ordered[j].Sequence
		}
		if !ordered[i].At.Equal(ordered[j].At) {
			return ordered[i].At.Before(ordered[j].At)
		}
		return ordered[i].ID < ordered[j].ID
	})
	sessions := []attendanceentity.Session{}
	state := attendanceentity.State{}
	for _, event := range ordered {
		if event.ResultCode != "" && !AcceptedCode(event.ResultCode) {
			continue
		}
		next, outcome := ApplyEvent(state, event)
		if !outcome.Accepted || (event.ResultCode != "" && outcome.Code != event.ResultCode) {
			issues = append(issues, Issue{event.ID, "replay_" + outcome.Code})
			conflict = true
			continue
		}
		state = next
		if outcome.Code == "clocked_out" {
			sessions = append(sessions, outcome.Session)
		}
	}
	if state.Active != nil {
		open := *state.Active
		if !state.RestStartedAt.IsZero() {
			open.Breaks = append(append([]attendancevalueobject.Interval(nil), open.Breaks...), attendancevalueobject.Interval{Start: state.RestStartedAt})
		}
		sessions = append(sessions, open)
	}

	if conflict {
		for i := range sessions {
			sessions[i].RestConfirmed = false
			sessions[i].Invalid = true
		}
	}
	return sessions, issues
}

// Calculate clips complete sessions to [from,to), separates confirmed and estimated totals, and
// keeps sub-minute precision until the final total. Invalid sessions never accrue work time.
func Calculate(sessions []attendanceentity.Session, from, to time.Time) Result {
	r := Result{Version: Version}
	if !to.After(from) {
		r.Issues = append(r.Issues, Issue{Code: "invalid_period"})
		return r
	}
	all := []attendancevalueobject.Interval{}
	estimated, attendance := []attendancevalueobject.Interval{}, []attendancevalueobject.Interval{}
	evidence := map[string]bool{}
	for _, s := range sessions {
		if !s.Start.IsZero() && !s.Start.Before(to) {
			continue
		}
		if !s.End.IsZero() && s.End.After(s.Start) && !s.End.After(from) {
			continue
		}
		for _, id := range s.EventIDs {
			if !evidence[id] {
				r.EventIDs = append(r.EventIDs, id)
				evidence[id] = true
			}
		}
		code := ""
		switch {
		case s.Start.IsZero():
			code = "missing_clock_in"
		case s.End.IsZero():
			code = "missing_clock_out"
		case !s.End.After(s.Start):
			code = "invalid_order"
		case s.Invalid:
			code = "invalid_evidence"
		}
		if code == "" {
			attendance = append(attendance, clipped([]attendancevalueobject.Interval{{Start: s.Start, End: s.End}}, from, to)...)
		}
		for _, b := range s.Breaks {
			if !b.ValidWithin(s.Start, s.End) {
				code = "invalid_break"
			}
		}
		if code != "" {
			r.Pending++
			r.Uncalculable++
			r.Issues = append(r.Issues, Issue{s.ID, code})
			continue
		}
		if !s.RestConfirmed {
			r.Pending++
			r.Unconfirmed++
			r.Issues = append(r.Issues, Issue{s.ID, "rest_unconfirmed"})
			estimated = append(estimated, clipped([]attendancevalueobject.Interval{{Start: s.Start, End: s.End}}, from, to)...)
			continue
		}
		breaks := union(append([]attendancevalueobject.Interval(nil), s.Breaks...))
		cursor := s.Start
		pieces := []attendancevalueobject.Interval{}
		for _, b := range breaks {
			if b.Start.After(cursor) {
				pieces = append(pieces, attendancevalueobject.Interval{Start: cursor, End: b.Start})
			}
			cursor = b.End
		}
		if cursor.Before(s.End) {
			pieces = append(pieces, attendancevalueobject.Interval{Start: cursor, End: s.End})
		}
		pieces = clipped(pieces, from, to)
		all = append(all, pieces...)
		estimated = append(estimated, pieces...)
	}
	// Overlap is retained as an anomaly and counted once, never double paid.
	sort.Slice(estimated, func(i, j int) bool { return estimated[i].Start.Before(estimated[j].Start) })
	var last time.Time
	for _, p := range estimated {
		if p.Start.Before(last) {
			r.Issues = append(r.Issues, Issue{Code: "overlapping_work"})
		}
		if p.End.After(last) {
			last = p.End
		}
	}
	r.Intervals = union(all)
	for _, p := range r.Intervals {
		r.Duration += p.End.Sub(p.Start)
	}
	r.ConfirmedNetMinutes = int64(r.Duration / time.Minute)
	r.EstimatedTotalMinutes = totalMinutes(estimated)
	r.AttendanceMinutes = totalMinutes(attendance)
	return r
}

func clipped(items []attendancevalueobject.Interval, from, to time.Time) []attendancevalueobject.Interval {
	result := []attendancevalueobject.Interval{}
	for _, p := range items {
		if p.Start.Before(from) {
			p.Start = from
		}
		if p.End.After(to) {
			p.End = to
		}
		if p.End.After(p.Start) {
			result = append(result, p)
		}
	}
	return result
}

func totalMinutes(items []attendancevalueobject.Interval) int64 {
	var duration time.Duration
	for _, p := range union(items) {
		duration += p.End.Sub(p.Start)
	}
	return int64(duration / time.Minute)
}

func union(items []attendancevalueobject.Interval) []attendancevalueobject.Interval {
	sort.Slice(items, func(i, j int) bool { return items[i].Start.Before(items[j].Start) })
	out := []attendancevalueobject.Interval{}
	for _, p := range items {
		n := len(out)
		if n == 0 || p.Start.After(out[n-1].End) {
			out = append(out, p)
		} else if p.End.After(out[n-1].End) {
			out[n-1].End = p.End
		}
	}
	return out
}
