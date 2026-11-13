package httpapi

import (
	"net/http"

	"github.com/963sup/line-mini-app-bot/internal/httpapi/generated"
)

type contractServer struct {
	IdentityServer
	AttendanceServer
	RepositoryServer
	IssueServer
	DiscussionServer
	LabelServer
	MilestoneServer
	ProjectServer
	StatisticsServer
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

type LabelServer interface {
	ListLabels(w http.ResponseWriter, r *http.Request, repositoryId string, params generated.ListLabelsParams)
	CreateLabel(w http.ResponseWriter, r *http.Request, repositoryId string, params generated.CreateLabelParams)
}

type IssueServer interface {
	ListIssues(w http.ResponseWriter, r *http.Request, repositoryId string, params generated.ListIssuesParams)
	CreateIssue(w http.ResponseWriter, r *http.Request, repositoryId string, params generated.CreateIssueParams)
	GetIssue(w http.ResponseWriter, r *http.Request, repositoryId string, id string)
	UpdateIssue(w http.ResponseWriter, r *http.Request, repositoryId string, id string, params generated.UpdateIssueParams)
}

type DiscussionServer interface {
	ListDiscussions(w http.ResponseWriter, r *http.Request, repositoryId string, params generated.ListDiscussionsParams)
	CreateDiscussion(w http.ResponseWriter, r *http.Request, repositoryId string, params generated.CreateDiscussionParams)
	GetDiscussion(w http.ResponseWriter, r *http.Request, repositoryId string, id string)
	UpdateDiscussion(w http.ResponseWriter, r *http.Request, repositoryId string, id string, params generated.UpdateDiscussionParams)
	ListDiscussionComments(w http.ResponseWriter, r *http.Request, repositoryId string, id string, params generated.ListDiscussionCommentsParams)
	CreateDiscussionComment(w http.ResponseWriter, r *http.Request, repositoryId string, id string, params generated.CreateDiscussionCommentParams)
	UpdateDiscussionComment(w http.ResponseWriter, r *http.Request, repositoryId string, id string, commentId string, params generated.UpdateDiscussionCommentParams)
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

type MilestoneServer interface {
	ListMilestones(w http.ResponseWriter, r *http.Request, repositoryId string, params generated.ListMilestonesParams)
	CreateMilestone(w http.ResponseWriter, r *http.Request, repositoryId string, params generated.CreateMilestoneParams)
	GetMilestone(w http.ResponseWriter, r *http.Request, repositoryId string, id string)
	UpdateMilestone(w http.ResponseWriter, r *http.Request, repositoryId string, id string, params generated.UpdateMilestoneParams)
}
type ProjectServer interface {
	ListProjects(w http.ResponseWriter, r *http.Request, repositoryId string, params generated.ListProjectsParams)
	CreateProject(w http.ResponseWriter, r *http.Request, repositoryId string, params generated.CreateProjectParams)
	GetProject(w http.ResponseWriter, r *http.Request, repositoryId string, id string)
	UpdateProject(w http.ResponseWriter, r *http.Request, repositoryId string, id string, params generated.UpdateProjectParams)
	ListProjectItems(w http.ResponseWriter, r *http.Request, repositoryId string, id string, params generated.ListProjectItemsParams)
	AddProjectItem(w http.ResponseWriter, r *http.Request, repositoryId string, id string, params generated.AddProjectItemParams)
	RemoveProjectItem(w http.ResponseWriter, r *http.Request, repositoryId string, id string, itemId string, params generated.RemoveProjectItemParams)
}
type StatisticsServer interface {
	GetRepositoryStatistics(w http.ResponseWriter, r *http.Request, repositoryId string)
}
