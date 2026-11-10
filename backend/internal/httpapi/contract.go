package httpapi

import (
	"net/http"

	"github.com/963sup/line-mini-app-bot/internal/httpapi/generated"
)

type contractServer struct {
	identity   http.Handler
	attendance http.Handler
}

var _ generated.ServerInterface = contractServer{}

func (s contractServer) GetMiniAppConfig(w http.ResponseWriter, r *http.Request) {
	s.identity.ServeHTTP(w, r)
}

func (s contractServer) GetCurrentUser(w http.ResponseWriter, r *http.Request) {
	s.identity.ServeHTTP(w, r)
}

func (s contractServer) ProposeAttendanceCorrection(w http.ResponseWriter, r *http.Request, params generated.ProposeAttendanceCorrectionParams) {
	s.attendance.ServeHTTP(w, r)
}

func (s contractServer) GetAttendanceCorrectionStatus(w http.ResponseWriter, r *http.Request, id string) {
	s.attendance.ServeHTTP(w, r)
}

func (s contractServer) ApproveAttendanceCorrection(w http.ResponseWriter, r *http.Request, id string, _ generated.ApproveAttendanceCorrectionParams) {
	s.attendance.ServeHTTP(w, r)
}

func (s contractServer) GetAttendanceCorrectionReview(w http.ResponseWriter, r *http.Request, id string) {
	s.attendance.ServeHTTP(w, r)
}

func (s contractServer) ListAttendanceGroups(w http.ResponseWriter, r *http.Request) {
	s.attendance.ServeHTTP(w, r)
}

func (s contractServer) ListAttendanceGroupPage(w http.ResponseWriter, r *http.Request, params generated.ListAttendanceGroupPageParams) {
	s.attendance.ServeHTTP(w, r)
}

func (s contractServer) ExchangeAttendanceLaunch(w http.ResponseWriter, r *http.Request, _ generated.ExchangeAttendanceLaunchParams) {
	s.attendance.ServeHTTP(w, r)
}

func (s contractServer) GetAttendanceMonthlyReport(w http.ResponseWriter, r *http.Request, params generated.GetAttendanceMonthlyReportParams) {
	s.attendance.ServeHTTP(w, r)
}

func (s contractServer) GetAttendanceMonthlySummary(w http.ResponseWriter, r *http.Request, params generated.GetAttendanceMonthlySummaryParams) {
	s.attendance.ServeHTTP(w, r)
}

func (s contractServer) SubmitAttendancePunch(w http.ResponseWriter, r *http.Request, _ generated.SubmitAttendancePunchParams) {
	s.attendance.ServeHTTP(w, r)
}

func (s contractServer) ListAttendanceRecords(w http.ResponseWriter, r *http.Request, params generated.ListAttendanceRecordsParams) {
	s.attendance.ServeHTTP(w, r)
}

func (s contractServer) GetAttendanceRecord(w http.ResponseWriter, r *http.Request, id string, params generated.GetAttendanceRecordParams) {
	s.attendance.ServeHTTP(w, r)
}

func (s contractServer) CreateSession(w http.ResponseWriter, r *http.Request, _ generated.CreateSessionParams) {
	s.identity.ServeHTTP(w, r)
}

func (s contractServer) DeleteSession(w http.ResponseWriter, r *http.Request, _ generated.DeleteSessionParams) {
	s.identity.ServeHTTP(w, r)
}

func (s contractServer) RefreshSession(w http.ResponseWriter, r *http.Request, _ generated.RefreshSessionParams) {
	s.identity.ServeHTTP(w, r)
}
