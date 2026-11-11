package attendancehttp

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/963sup/line-mini-app-bot/internal/business/attendance/punch"
	recordmodel "github.com/963sup/line-mini-app-bot/internal/business/attendance/records"
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
	records            []recordmodel.Record
	punchCalls         int
}

func (s *attendanceStub) SubmitPunch(context.Context, string, string, recordmodel.Position, time.Time) (punch.Result, error) {
	return punch.Result{}, s.err
}

func (s *attendanceStub) Groups(_ context.Context, user string) ([]string, error) {
	s.queriedUser = user
	return s.groups, s.err
}

func (s *attendanceStub) RecentGroup(_ context.Context, user, group string) ([]recordmodel.Record, error) {
	s.queriedUser = user
	s.queriedGroup = group
	return nil, s.err
}

func (s *attendanceStub) Punch(_ context.Context, group, user, event, action string, now time.Time, _ int64) (punch.Result, error) {
	s.punchCalls++
	s.calls++
	s.group, s.user, s.event = group, user, event
	s.action = action
	return punch.Result{Code: "clocked_in", Record: recordmodel.Record{ClockInAt: now}}, s.err
}

func (s *attendanceStub) Recent(_ context.Context, user string) ([]recordmodel.Record, error) {
	s.queriedUser = user
	return s.records, s.err
}

func (s *attendanceStub) RecordsOverlappingMonth(_ context.Context, user, group string, _ time.Time) ([]recordmodel.Record, error) {
	s.queriedUser = user
	s.queriedGroup = group
	return s.records, s.err
}
