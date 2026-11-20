package service

import (
	"time"

	attendanceentity "github.com/963sup/line-mini-app-bot/internal/business/attendance/domain/entity"
	attendanceevent "github.com/963sup/line-mini-app-bot/internal/business/attendance/domain/event"
	attendancevalueobject "github.com/963sup/line-mini-app-bot/internal/business/attendance/domain/valueobject"
)

type Outcome struct {
	Code     string
	Accepted bool
	Session  attendanceentity.Session
}

func ApplyEvent(state attendanceentity.State, event attendanceevent.Event) (attendanceentity.State, Outcome) {
	if event.ID == "" || event.At.IsZero() {
		return state, Outcome{Code: "invalid_event"}
	}
	if !state.LastAt.IsZero() && event.At.Before(state.LastAt) {
		return state, Outcome{Code: "stale_event"}
	}
	next := state
	if state.Active != nil {
		copy := *state.Active
		copy.Breaks = append([]attendancevalueobject.Interval(nil), copy.Breaks...)
		copy.EventIDs = append([]string(nil), copy.EventIDs...)
		next.Active = &copy
	}
	result := Outcome{}
	switch event.Kind {
	case "clock-in":
		if next.Active != nil {
			return state, Outcome{Code: "already_in", Session: *next.Active}
		}
		next.Active = &attendanceentity.Session{ID: event.ID, Start: event.At}
		result.Code = "clocked_in"
	case "break-start":
		if next.Active == nil {
			return state, Outcome{Code: "not_in"}
		}
		if !next.RestStartedAt.IsZero() {
			return state, Outcome{Code: "already_resting", Session: *next.Active}
		}
		next.RestStartedAt = event.At
		next.Active.RestConfirmed = false
		result.Code = "break_started"
	case "break-end":
		if next.Active == nil {
			return state, Outcome{Code: "not_in"}
		}
		if next.RestStartedAt.IsZero() {
			return state, Outcome{Code: "not_resting", Session: *next.Active}
		}
		if !event.At.After(next.RestStartedAt) {
			return state, Outcome{Code: "invalid_break", Session: *next.Active}
		}
		next.Active.Breaks = append(next.Active.Breaks, attendancevalueobject.Interval{Start: next.RestStartedAt, End: event.At})
		next.RestStartedAt = time.Time{}
		next.Active.RestConfirmed = false
		result.Code = "break_ended"
	case "confirm-rest":
		if next.Active == nil {
			return state, Outcome{Code: "not_in"}
		}
		if !next.RestStartedAt.IsZero() {
			return state, Outcome{Code: "break_incomplete", Session: *next.Active}
		}
		next.Active.RestConfirmed = true
		result.Code = "rest_confirmed"
	case "clock-out":
		if next.Active == nil {
			return state, Outcome{Code: "not_in"}
		}
		if !event.At.After(next.Active.Start) {
			return state, Outcome{Code: "invalid_order", Session: *next.Active}
		}
		next.Active.EventIDs = append(next.Active.EventIDs, event.ID)
		next.Active.End = event.At
		if !next.RestStartedAt.IsZero() {
			next.Active.Breaks = append(next.Active.Breaks, attendancevalueobject.Interval{Start: next.RestStartedAt})
			next.RestStartedAt = time.Time{}
		}
		result.Session = *next.Active
		next.Active = nil
		result.Code = "clocked_out"
	default:
		return state, Outcome{Code: "unknown_event"}
	}
	next.LastAt = event.At
	result.Accepted = true
	if next.Active != nil {
		next.Active.EventIDs = append(next.Active.EventIDs, event.ID)
		result.Session = *next.Active
	}
	return next, result
}

func AcceptedCode(code string) bool {
	switch code {
	case "clocked_in", "clocked_out", "break_started", "break_ended", "rest_confirmed":
		return true
	}
	return false
}
