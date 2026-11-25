package attendancehttp

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	attendancedto "github.com/963sup/line-mini-app-bot/internal/business/attendance/application/dto"
	attendanceport "github.com/963sup/line-mini-app-bot/internal/business/attendance/application/port"
	attendanceentity "github.com/963sup/line-mini-app-bot/internal/business/attendance/domain/entity"
	attendancevalueobject "github.com/963sup/line-mini-app-bot/internal/business/attendance/domain/valueobject"
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
	attendanceport.LeaveStore
	attendanceport.DailyHeadcountReader
	attendanceport.RecordQueries
	attendanceport.CorrectionSubmitter
	attendanceport.IdempotentCorrectionStore
	attendanceport.CorrectionReader
	attendanceport.ReviewerGroupReader
	attendanceport.CorrectionPageReader
	attendanceport.CorrectionReviewRunner
	attendanceport.PunchResolver
	attendanceport.LaunchStore
	calls              int
	err                error
	group, user, event string
	action             string
	queriedUser        string
	queriedGroup       string
	groups             []string
	records            []attendanceentity.Record
	punchCalls         int
}

func (s *attendanceStub) SubmitPunch(context.Context, string, string, attendancevalueobject.Position, time.Time) (attendancedto.PunchResult, error) {
	return attendancedto.PunchResult{}, s.err
}

func (s *attendanceStub) Groups(_ context.Context, user string) ([]string, error) {
	s.queriedUser = user
	return s.groups, s.err
}

func (s *attendanceStub) RecentGroup(_ context.Context, user, group string) ([]attendanceentity.Record, error) {
	s.queriedUser = user
	s.queriedGroup = group
	return nil, s.err
}

func (s *attendanceStub) Punch(_ context.Context, group, user, event, action string, now time.Time, _ int64) (attendancedto.PunchResult, error) {
	s.punchCalls++
	s.calls++
	s.group, s.user, s.event = group, user, event
	s.action = action
	return attendancedto.PunchResult{Code: "clocked_in", Record: attendanceentity.Record{ClockInAt: now}}, s.err
}

func (s *attendanceStub) Recent(_ context.Context, user string) ([]attendanceentity.Record, error) {
	s.queriedUser = user
	return s.records, s.err
}

func (s *attendanceStub) RecordsOverlappingMonth(_ context.Context, user, group string, _ time.Time) ([]attendanceentity.Record, error) {
	s.queriedUser = user
	s.queriedGroup = group
	return s.records, s.err
}

func TestCorrectionActionsForAuthorizedScope(t *testing.T) {
	for _, tc := range []struct {
		status                  string
		approve, reject, cancel bool
	}{
		{"pending", true, true, true}, {"stale", false, true, true}, {"approved", false, false, false}, {"rejected", false, false, false}, {"cancelled", false, false, false}, {"invalid", false, false, false}, {"", false, false, false},
	} {
		details := attendancedto.CorrectionDetails{Status: tc.status}
		reviewer := correctionDetailsView(details, true)
		own := correctionDetailsView(details, false)
		if reviewer.CanApprove != tc.approve || reviewer.CanReject != tc.reject || reviewer.CanCancel || own.CanReject || own.CanApprove || own.CanCancel != tc.cancel {
			t.Errorf("wrong actions for %q: review=%+v own=%+v", tc.status, reviewer, own)
		}
	}
}
