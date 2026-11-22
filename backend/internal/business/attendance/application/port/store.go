package port

import (
	"context"
	"time"

	attendancedto "github.com/963sup/line-mini-app-bot/internal/business/attendance/application/dto"
	attendanceentity "github.com/963sup/line-mini-app-bot/internal/business/attendance/domain/entity"
)

type Store interface {
	Punch(ctx context.Context, groupID, userID, eventID, action string, now time.Time, eventTimeMillis int64) (attendancedto.PunchResult, error)
	Recent(ctx context.Context, userID string) ([]attendanceentity.Record, error)
}

// GroupHistoryReader reads only the authenticated user's group history.
type GroupHistoryReader interface {
	Groups(ctx context.Context, userID string) ([]string, error)
	RecentGroup(ctx context.Context, userID, groupID string) ([]attendanceentity.Record, error)
}

type LeaveStore interface {
	CreateLeaveLaunch(ctx context.Context, groupID string) (string, error)
	LeaveGroup(ctx context.Context, token string) (string, error)
	SubmitLeave(ctx context.Context, token, userID, retryKey string, proposal attendanceentity.Leave) (attendanceentity.Leave, error)
	RecentLeaves(ctx context.Context, userID string) ([]attendanceentity.Leave, error)
}

type BoardStore interface {
	GroupBoard(ctx context.Context, groupID string, start, now time.Time) (attendancedto.Board, error)
	CurrentShift(ctx context.Context, groupID, userID string) (*attendanceentity.Record, error)
}

type Membership interface {
	IsMember(context.Context, string, string) (bool, error)
}

type RecordQueries interface {
	GroupsPage(ctx context.Context, userID, cursor string, pageSize int) (attendancedto.GroupPage, error)
	RecordsPage(ctx context.Context, userID, groupID string, month time.Time, cursor string, pageSize int) (attendancedto.RecordPage, error)
	Record(ctx context.Context, userID, groupID, recordID string) (attendanceentity.Record, error)
	MonthlyReader
}

type CorrectionStore interface {
	ProposeCorrection(context.Context, string, attendanceentity.Correction, time.Time) (string, error)
	CorrectionReviewRunner
}

type IdempotentCorrectionStore interface {
	ProposeCorrectionOnce(context.Context, string, string, attendanceentity.Correction, time.Time) (string, error)
}

type CorrectionReader interface {
	CorrectionDetails(context.Context, string, string, bool) (attendancedto.CorrectionDetails, error)
}

type ReviewerGroupReader interface {
	ReviewerGroups(context.Context, string) ([]string, error)
}

type CorrectionPageReader interface {
	CorrectionsPage(context.Context, string, string, string, bool, int) (attendancedto.CorrectionPage, error)
}
