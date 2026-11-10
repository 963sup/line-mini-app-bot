package attendancehttp

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/963sup/line-mini-app-bot/internal/attendance"
)

func TestAttendanceBrowserRequiresAuthentication(t *testing.T) {
	t.Setenv("LINE_MINI_APP_CHANNEL_ID", "123")
	mux := http.NewServeMux()
	RegisterAttendanceAPI(mux, &Attendance{store: &attendanceStub{}}, testAuthenticator(&http.Client{}))
	r := httptest.NewRequest(http.MethodPost, "/api/v1/attendance/punches", strings.NewReader(`{}`))
	w := httptest.NewRecorder()
	mux.ServeHTTP(w, r)
	if w.Code != http.StatusUnauthorized {
		t.Fatal("browser punch must require LINE authentication")
	}
}

type attendanceStub struct {
	calls              int
	err                error
	group, user, event string
	action             string
	queriedUser        string
	queriedGroup       string
	groups             []string
	records            []attendance.Record
	punchCalls         int
}

func (s *attendanceStub) Submit(context.Context, string, string, attendance.Position, time.Time) (attendance.Result, error) {
	return attendance.Result{}, s.err
}

func (s *attendanceStub) Groups(_ context.Context, user string) ([]string, error) {
	s.queriedUser = user
	return s.groups, s.err
}

func (s *attendanceStub) RecentGroup(_ context.Context, user, group string) ([]attendance.Record, error) {
	s.queriedUser = user
	s.queriedGroup = group
	return nil, s.err
}

func (s *attendanceStub) Punch(_ context.Context, group, user, event, action string, now time.Time, _ int64) (attendance.Result, error) {
	s.punchCalls++
	s.calls++
	s.group, s.user, s.event = group, user, event
	s.action = action
	return attendance.Result{Code: "clocked_in", Record: attendance.Record{ClockInAt: now}}, s.err
}

func (s *attendanceStub) Recent(_ context.Context, user string) ([]attendance.Record, error) {
	s.queriedUser = user
	return s.records, s.err
}

func (s *attendanceStub) MonthGroup(_ context.Context, user, group string, _ time.Time) ([]attendance.Record, error) {
	s.queriedUser = user
	s.queriedGroup = group
	return s.records, s.err
}
