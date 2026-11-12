package httpapi

import (
	"net/http"

	"github.com/963sup/line-mini-app-bot/internal/httpapi/generated"
)

type contractServer struct {
	identity     IdentityServer
	attendance   AttendanceServer
	repositories RepositoryServer
}

var _ generated.ServerInterface = contractServer{}

func (s contractServer) ListRepositories(w http.ResponseWriter, r *http.Request, params generated.ListRepositoriesParams) {
	s.repositories.ListRepositories(w, r, params)
}

func (s contractServer) GetRepository(w http.ResponseWriter, r *http.Request, id string) {
	s.repositories.GetRepository(w, r, id)
}

func (s contractServer) GetMiniAppConfig(w http.ResponseWriter, r *http.Request) {
	s.identity.GetMiniAppConfig(w, r)
}

func (s contractServer) GetCurrentUser(w http.ResponseWriter, r *http.Request) {
	s.identity.GetCurrentUser(w, r)
}

func (s contractServer) ProposeAttendanceCorrection(w http.ResponseWriter, r *http.Request, params generated.ProposeAttendanceCorrectionParams) {
	s.attendance.ProposeAttendanceCorrection(w, r, params)
}

func (s contractServer) GetAttendanceCorrectionStatus(w http.ResponseWriter, r *http.Request, id string) {
	s.attendance.GetAttendanceCorrectionStatus(w, r, id)
}

func (s contractServer) ApproveAttendanceCorrection(w http.ResponseWriter, r *http.Request, id string, params generated.ApproveAttendanceCorrectionParams) {
	s.attendance.ApproveAttendanceCorrection(w, r, id, params)
}

func (s contractServer) GetAttendanceCorrectionReview(w http.ResponseWriter, r *http.Request, id string) {
	s.attendance.GetAttendanceCorrectionReview(w, r, id)
}

func (s contractServer) ListAttendanceGroups(w http.ResponseWriter, r *http.Request) {
	s.attendance.ListAttendanceGroups(w, r)
}

func (s contractServer) ListAttendanceGroupPage(w http.ResponseWriter, r *http.Request, params generated.ListAttendanceGroupPageParams) {
	s.attendance.ListAttendanceGroupPage(w, r, params)
}

func (s contractServer) ExchangeAttendanceLaunch(w http.ResponseWriter, r *http.Request, params generated.ExchangeAttendanceLaunchParams) {
	s.attendance.ExchangeAttendanceLaunch(w, r, params)
}

func (s contractServer) GetAttendanceMonthlyReport(w http.ResponseWriter, r *http.Request, params generated.GetAttendanceMonthlyReportParams) {
	s.attendance.GetAttendanceMonthlyReport(w, r, params)
}

func (s contractServer) GetAttendanceMonthlySummary(w http.ResponseWriter, r *http.Request, params generated.GetAttendanceMonthlySummaryParams) {
	s.attendance.GetAttendanceMonthlySummary(w, r, params)
}

func (s contractServer) SubmitAttendancePunch(w http.ResponseWriter, r *http.Request, params generated.SubmitAttendancePunchParams) {
	s.attendance.SubmitAttendancePunch(w, r, params)
}

func (s contractServer) ListAttendanceRecords(w http.ResponseWriter, r *http.Request, params generated.ListAttendanceRecordsParams) {
	s.attendance.ListAttendanceRecords(w, r, params)
}

func (s contractServer) GetAttendanceRecord(w http.ResponseWriter, r *http.Request, id string, params generated.GetAttendanceRecordParams) {
	s.attendance.GetAttendanceRecord(w, r, id, params)
}

func (s contractServer) CreateSession(w http.ResponseWriter, r *http.Request, params generated.CreateSessionParams) {
	s.identity.CreateSession(w, r, params)
}

func (s contractServer) DeleteSession(w http.ResponseWriter, r *http.Request, params generated.DeleteSessionParams) {
	s.identity.DeleteSession(w, r, params)
}

func (s contractServer) RefreshSession(w http.ResponseWriter, r *http.Request, params generated.RefreshSessionParams) {
	s.identity.RefreshSession(w, r, params)
}

func (s contractServer) ListOwnLeaveRequests(w http.ResponseWriter, r *http.Request) {
	s.attendance.ListOwnLeaveRequests(w, r)
}

func (s contractServer) SubmitLeaveRequest(w http.ResponseWriter, r *http.Request, params generated.SubmitLeaveRequestParams) {
	s.attendance.SubmitLeaveRequest(w, r, params)
}

func (s contractServer) PreflightAttendancePunch(w http.ResponseWriter, r *http.Request, params generated.PreflightAttendancePunchParams) {
	s.attendance.PreflightAttendancePunch(w, r, params)
}

func (s contractServer) RejectAttendanceCorrection(w http.ResponseWriter, r *http.Request, id string, params generated.RejectAttendanceCorrectionParams) {
	s.attendance.RejectAttendanceCorrection(w, r, id, params)
}

func (s contractServer) CancelAttendanceCorrection(w http.ResponseWriter, r *http.Request, id string, params generated.CancelAttendanceCorrectionParams) {
	s.attendance.CancelAttendanceCorrection(w, r, id, params)
}

func (s contractServer) ListAttendanceCorrections(w http.ResponseWriter, r *http.Request, params generated.ListAttendanceCorrectionsParams) {
	s.attendance.ListAttendanceCorrections(w, r, params)
}

func (s contractServer) ListAttendanceReviewerGroups(w http.ResponseWriter, r *http.Request) {
	s.attendance.ListAttendanceReviewerGroups(w, r)
}

// IdentityServer is the generated operation subset owned by this capability.
type IdentityServer interface {
	GetMiniAppConfig(w http.ResponseWriter, r *http.Request)
	GetCurrentUser(w http.ResponseWriter, r *http.Request)
	CreateSession(w http.ResponseWriter, r *http.Request, params generated.CreateSessionParams)
	DeleteSession(w http.ResponseWriter, r *http.Request, params generated.DeleteSessionParams)
	RefreshSession(w http.ResponseWriter, r *http.Request, params generated.RefreshSessionParams)
}

// RepositoryServer is the generated operation subset owned by this capability.
type RepositoryServer interface {
	ListRepositories(w http.ResponseWriter, r *http.Request, params generated.ListRepositoriesParams)
	GetRepository(w http.ResponseWriter, r *http.Request, id string)
}

// AttendanceServer is the generated operation subset owned by this capability.
type AttendanceServer interface {
	ProposeAttendanceCorrection(w http.ResponseWriter, r *http.Request, params generated.ProposeAttendanceCorrectionParams)
	GetAttendanceCorrectionStatus(w http.ResponseWriter, r *http.Request, id string)
	ApproveAttendanceCorrection(w http.ResponseWriter, r *http.Request, id string, params generated.ApproveAttendanceCorrectionParams)
	GetAttendanceCorrectionReview(w http.ResponseWriter, r *http.Request, id string)
	ListAttendanceGroups(w http.ResponseWriter, r *http.Request)
	ListAttendanceGroupPage(w http.ResponseWriter, r *http.Request, params generated.ListAttendanceGroupPageParams)
	ExchangeAttendanceLaunch(w http.ResponseWriter, r *http.Request, params generated.ExchangeAttendanceLaunchParams)
	GetAttendanceMonthlyReport(w http.ResponseWriter, r *http.Request, params generated.GetAttendanceMonthlyReportParams)
	GetAttendanceMonthlySummary(w http.ResponseWriter, r *http.Request, params generated.GetAttendanceMonthlySummaryParams)
	SubmitAttendancePunch(w http.ResponseWriter, r *http.Request, params generated.SubmitAttendancePunchParams)
	ListAttendanceRecords(w http.ResponseWriter, r *http.Request, params generated.ListAttendanceRecordsParams)
	GetAttendanceRecord(w http.ResponseWriter, r *http.Request, id string, params generated.GetAttendanceRecordParams)
	ListOwnLeaveRequests(w http.ResponseWriter, r *http.Request)
	SubmitLeaveRequest(w http.ResponseWriter, r *http.Request, params generated.SubmitLeaveRequestParams)
	PreflightAttendancePunch(w http.ResponseWriter, r *http.Request, params generated.PreflightAttendancePunchParams)
	RejectAttendanceCorrection(w http.ResponseWriter, r *http.Request, id string, params generated.RejectAttendanceCorrectionParams)
	CancelAttendanceCorrection(w http.ResponseWriter, r *http.Request, id string, params generated.CancelAttendanceCorrectionParams)
	ListAttendanceCorrections(w http.ResponseWriter, r *http.Request, params generated.ListAttendanceCorrectionsParams)
	ListAttendanceReviewerGroups(w http.ResponseWriter, r *http.Request)
}
