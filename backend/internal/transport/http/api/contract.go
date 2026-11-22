package httpapi

import (
	"net/http"

	"github.com/963sup/line-mini-app-bot/internal/transport/http/api/generated"
)

type EnterpriseServer interface {
	AttachEnterpriseOrganization(w http.ResponseWriter, r *http.Request, enterpriseId string)
	AddEnterpriseOwner(w http.ResponseWriter, r *http.Request, enterpriseId string)
	GetEnterprise(w http.ResponseWriter, r *http.Request, enterpriseId string)
	CreateEnterprise(w http.ResponseWriter, r *http.Request)
}

type TeamServer interface {
	RemoveEnterpriseTeam(w http.ResponseWriter, r *http.Request, organizationId string, teamId string)
	AssignEnterpriseTeam(w http.ResponseWriter, r *http.Request, organizationId string, teamId string)
	ListOrganizationTeams(w http.ResponseWriter, r *http.Request, organizationId string)
	SetTeamManagement(w http.ResponseWriter, r *http.Request, teamId string)
}

type OrganizationServer interface {
	TransferOrganization(w http.ResponseWriter, r *http.Request, organizationId string)
	CreateOrganization(w http.ResponseWriter, r *http.Request)
	GetOrganizationContext(w http.ResponseWriter, r *http.Request)
	SelectOrganization(w http.ResponseWriter, r *http.Request)
}

type MembershipServer interface {
	ListOrganizationMembers(w http.ResponseWriter, r *http.Request, organizationId string)
	LeaveOrganization(w http.ResponseWriter, r *http.Request, organizationId string, params generated.LeaveOrganizationParams)
	RemoveOrganizationMember(w http.ResponseWriter, r *http.Request, organizationId string, memberId string, params generated.RemoveOrganizationMemberParams)
}

type PopplerServer interface {
	ExtractPopplerText(w http.ResponseWriter, r *http.Request, params generated.ExtractPopplerTextParams)
}

type contractServer struct {
	PopplerServer
	EnterpriseServer
	TeamServer
	OrganizationServer
	MembershipServer
	JournalServer
	IdentityServer
	AttendanceServer
	RepositoryServer
	IssueServer
	ObjectServer
	RepositorySpaceServer
	DiscussionServer
	WikiServer
	LabelServer
	MilestoneServer
	ProjectServer
	StatisticsServer
	GroupServer
	OwnerProjectServer
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
	RenameLabel(w http.ResponseWriter, r *http.Request, repositoryId string, id string)
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
	GetAttendanceDailyHeadcount(w http.ResponseWriter, r *http.Request, params generated.GetAttendanceDailyHeadcountParams)
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

type GroupServer interface {
	ListGroups(w http.ResponseWriter, r *http.Request, params generated.ListGroupsParams)
	GetGroup(w http.ResponseWriter, r *http.Request, id string)
}
type OwnerProjectServer interface {
	GetProjectReferenceState(w http.ResponseWriter, r *http.Request, ownerId string, id string, params generated.GetProjectReferenceStateParams)
	ListInspectionSubmissions(w http.ResponseWriter, r *http.Request, ownerId string, id string, itemId string)
	SubmitInspection(w http.ResponseWriter, r *http.Request, ownerId string, id string, itemId string, params generated.SubmitInspectionParams)
	ListProjectReferences(w http.ResponseWriter, r *http.Request, ownerId string, id string, params generated.ListProjectReferencesParams)
	AddProjectReference(w http.ResponseWriter, r *http.Request, ownerId string, id string, params generated.AddProjectReferenceParams)
	RemoveProjectReference(w http.ResponseWriter, r *http.Request, ownerId string, id string, itemId string, params generated.RemoveProjectReferenceParams)
	ListOwnerProjects(w http.ResponseWriter, r *http.Request, ownerId string, params generated.ListOwnerProjectsParams)
	CreateOwnerProject(w http.ResponseWriter, r *http.Request, ownerId string, params generated.CreateOwnerProjectParams)
	GetOwnerProject(w http.ResponseWriter, r *http.Request, ownerId string, id string)
	UpdateOwnerProject(w http.ResponseWriter, r *http.Request, ownerId string, id string, params generated.UpdateOwnerProjectParams)
	ListOwnerProjectItems(w http.ResponseWriter, r *http.Request, ownerId string, id string, params generated.ListOwnerProjectItemsParams)
	AddOwnerProjectItem(w http.ResponseWriter, r *http.Request, ownerId string, id string, params generated.AddOwnerProjectItemParams)
	RemoveOwnerProjectItem(w http.ResponseWriter, r *http.Request, ownerId string, id string, itemId string, params generated.RemoveOwnerProjectItemParams)
}

type ObjectServer interface {
	ListPersonalObjectTemplates(w http.ResponseWriter, r *http.Request, params generated.ListPersonalObjectTemplatesParams)
	CreatePersonalObjectTemplate(w http.ResponseWriter, r *http.Request, params generated.CreatePersonalObjectTemplateParams)
	CreateObjectTemplate(w http.ResponseWriter, r *http.Request, repositoryId string, params generated.CreateObjectTemplateParams)
	ListObjectTemplates(w http.ResponseWriter, r *http.Request, repositoryId string, params generated.ListObjectTemplatesParams)
	GetObjectTemplate(w http.ResponseWriter, r *http.Request, repositoryId string, id string)
	InstantiateObjectTemplate(w http.ResponseWriter, r *http.Request, repositoryId string, id string, params generated.InstantiateObjectTemplateParams)
	ExtractObjectCandidates(w http.ResponseWriter, r *http.Request, repositoryId string, params generated.ExtractObjectCandidatesParams)
	ListObjects(w http.ResponseWriter, r *http.Request, repositoryId string, params generated.ListObjectsParams)
	CreateObject(w http.ResponseWriter, r *http.Request, repositoryId string, params generated.CreateObjectParams)
	GetObject(w http.ResponseWriter, r *http.Request, repositoryId string, id string)
	UpdateObject(w http.ResponseWriter, r *http.Request, repositoryId string, id string, params generated.UpdateObjectParams)
}

// RepositorySpaceServer owns real business spaces; GroupServer handles Project-owning LINE groups.
type RepositorySpaceServer interface {
	ListRepositorySpaces(w http.ResponseWriter, r *http.Request, repositoryId string, params generated.ListRepositorySpacesParams)
	CreateRepositorySpace(w http.ResponseWriter, r *http.Request, repositoryId string, params generated.CreateRepositorySpaceParams)
	GetRepositorySpace(w http.ResponseWriter, r *http.Request, repositoryId string, id string)
}

type JournalServer interface {
	ListJournals(w http.ResponseWriter, r *http.Request, params generated.ListJournalsParams)
	GetJournal(w http.ResponseWriter, r *http.Request, journalId string)
	CreateJournal(w http.ResponseWriter, r *http.Request, params generated.CreateJournalParams)
}

type WikiServer interface {
	ListWikiPages(w http.ResponseWriter, r *http.Request, repositoryId string, params generated.ListWikiPagesParams)
	CreateWikiPage(w http.ResponseWriter, r *http.Request, repositoryId string, params generated.CreateWikiPageParams)
	GetWikiPage(w http.ResponseWriter, r *http.Request, repositoryId string, id string)
	UpdateWikiPage(w http.ResponseWriter, r *http.Request, repositoryId string, id string)
	ListWikiRevisions(w http.ResponseWriter, r *http.Request, repositoryId string, id string, params generated.ListWikiRevisionsParams)
}
