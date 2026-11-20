package application

import (
	"context"
	"time"

	"github.com/963sup/line-mini-app-bot/internal/business/attendance/application/punch"
	"github.com/963sup/line-mini-app-bot/internal/business/attendance/domain/board"
	correctionsmodel "github.com/963sup/line-mini-app-bot/internal/business/attendance/domain/corrections"
	"github.com/963sup/line-mini-app-bot/internal/business/attendance/domain/leave"
	recordmodel "github.com/963sup/line-mini-app-bot/internal/business/attendance/domain/records"
)

type Store interface {
	Punch(ctx context.Context, groupID, userID, eventID, action string, now time.Time, eventTimeMillis int64) (punch.Result, error)
	Recent(ctx context.Context, userID string) ([]recordmodel.Record, error)
}

// GroupHistoryReader reads only the authenticated user's group history.
type GroupHistoryReader interface {
	Groups(ctx context.Context, userID string) ([]string, error)
	RecentGroup(ctx context.Context, userID, groupID string) ([]recordmodel.Record, error)
}

type LeaveStore interface {
	CreateLeaveLaunch(ctx context.Context, groupID string) (string, error)
	LeaveGroup(ctx context.Context, token string) (string, error)
	SubmitLeave(ctx context.Context, token, userID, retryKey string, proposal leave.Leave) (leave.Leave, error)
	RecentLeaves(ctx context.Context, userID string) ([]leave.Leave, error)
}

type BoardStore interface {
	GroupBoard(ctx context.Context, groupID string, start, now time.Time) (board.Board, error)
	CurrentShift(ctx context.Context, groupID, userID string) (*recordmodel.Record, error)
}

type Membership interface {
	IsMember(context.Context, string, string) (bool, error)
}

type RecordQueries interface {
	GroupsPage(ctx context.Context, userID, cursor string, pageSize int) (recordmodel.GroupPage, error)
	RecordsPage(ctx context.Context, userID, groupID string, month time.Time, cursor string, pageSize int) (recordmodel.RecordPage, error)
	Record(ctx context.Context, userID, groupID, recordID string) (recordmodel.Record, error)
	MonthlyReader
}

type CorrectionStore interface {
	ProposeCorrection(context.Context, string, correctionsmodel.Correction, time.Time) (string, error)
	ApproveCorrection(context.Context, string, string, time.Time) error
}

type IdempotentCorrectionStore interface {
	ProposeCorrectionOnce(context.Context, string, string, correctionsmodel.Correction, time.Time) (string, error)
}

type CorrectionReader interface {
	CorrectionDetails(context.Context, string, string, bool) (correctionsmodel.CorrectionDetails, error)
}

type ReviewerGroupReader interface {
	ReviewerGroups(context.Context, string) ([]string, error)
}

type CorrectionPageReader interface {
	CorrectionsPage(context.Context, string, string, string, bool, int) (correctionsmodel.CorrectionPage, error)
}

type CorrectionRejecter interface {
	RejectCorrection(context.Context, string, string, string, time.Time) error
}

type CorrectionCanceller interface {
	CancelCorrection(context.Context, string, string, time.Time) error
}
