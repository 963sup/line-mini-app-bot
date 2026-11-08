// Package worktime calculates recorded time without assuming schedules or pay rules.
package worktime

import (
	"sort"
	"time"
)

const Version = "worktime-v2"

type Event struct {
	ContractVersion   int
	RecordID, GroupID string
	ID                string
	Kind              string
	At                time.Time
	Source            string
	Sequence          int64
	ResultCode        string
}
type Interval struct{ Start, End time.Time }
type Session struct {
	ID            string
	Start, End    time.Time
	Breaks        []Interval
	RestConfirmed bool
	EventIDs      []string
}
type Issue struct{ SessionID, Code string }
type Result struct {
	EventIDs  []string
	Version   string
	Duration  time.Duration
	Minutes   int64
	Pending   int
	Issues    []Issue
	Intervals []Interval
}

// BuildIntervals preserves input, deduplicates event IDs and reports ambiguous sequences.
func BuildIntervals(events []Event) ([]Session, []Issue) {
	seen := map[string]Event{}
	ordered := []Event{}
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
	sessions := []Session{}
	state := State{}
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
			open.Breaks = append(append([]Interval(nil), open.Breaks...), Interval{Start: state.RestStartedAt})
		}
		sessions = append(sessions, open)
	}

	if conflict {
		for i := range sessions {
			sessions[i].RestConfirmed = false
		}
	}
	return sessions, issues
}

// Calculate clips complete confirmed sessions to [from,to), unions overlaps and
// keeps sub-minute precision until the final total. Invalid sessions never accrue time.
func Calculate(sessions []Session, from, to time.Time) Result {
	r := Result{Version: Version}
	if !to.After(from) {
		r.Issues = append(r.Issues, Issue{Code: "invalid_period"})
		return r
	}
	all := []Interval{}
	evidence := map[string]bool{}
	for _, s := range sessions {
		if !s.Start.IsZero() && !s.Start.Before(to) {
			continue
		}
		if !s.End.IsZero() && !s.End.After(from) && !s.End.Before(s.Start) {
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
		case !s.RestConfirmed:
			code = "rest_unconfirmed"
		}
		breaks := append([]Interval(nil), s.Breaks...)
		for _, b := range breaks {
			if b.Start.Before(s.Start) || b.End.After(s.End) || !b.End.After(b.Start) {
				code = "invalid_break"
			}
		}
		if code != "" {
			r.Pending++
			r.Issues = append(r.Issues, Issue{s.ID, code})
			continue
		}
		breaks = union(breaks)
		cursor := s.Start
		pieces := []Interval{}
		for _, b := range breaks {
			if b.Start.After(cursor) {
				pieces = append(pieces, Interval{cursor, b.Start})
			}
			cursor = b.End
		}
		if cursor.Before(s.End) {
			pieces = append(pieces, Interval{cursor, s.End})
		}
		for _, p := range pieces {
			if p.Start.Before(from) {
				p.Start = from
			}
			if p.End.After(to) {
				p.End = to
			}
			if p.End.After(p.Start) {
				all = append(all, p)
			}
		}
	}
	// Overlap is retained as an anomaly and counted once, never double paid.
	sorted := append([]Interval(nil), all...)
	sort.Slice(sorted, func(i, j int) bool { return sorted[i].Start.Before(sorted[j].Start) })
	var last time.Time
	for _, p := range sorted {
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
	r.Minutes = int64(r.Duration / time.Minute)
	return r
}
func union(items []Interval) []Interval {
	sort.Slice(items, func(i, j int) bool { return items[i].Start.Before(items[j].Start) })
	out := []Interval{}
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
