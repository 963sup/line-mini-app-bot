package httpapi

import (
	"net/http"

	"github.com/963sup/line-mini-app-bot/internal/httpapi/generated"
)

type contractServer struct {
	IdentityServer
	AttendanceServer
	RepositoryServer
}

var _ generated.ServerInterface = contractServer{}

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
