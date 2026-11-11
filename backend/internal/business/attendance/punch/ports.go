package punch

import (
	"context"
	"time"

	recordmodel "github.com/963sup/line-mini-app-bot/internal/business/attendance/records"
)

// PunchSubmitter records a located action bound to the authenticated user.
type PunchSubmitter interface {
	SubmitPunch(ctx context.Context, requestID, userID string, position recordmodel.Position, now time.Time) (Result, error)
}

// PunchResolver returns a receipt or seals a rejected attempt; it may write.
// needs_position leaves the attempt unconsumed for SubmitPunch.
type PunchResolver interface {
	ResolvePunchAttempt(ctx context.Context, requestID, userID string, now time.Time) (Result, error)
}

type LaunchStore interface {
	CreateLaunch(ctx context.Context, groupID, action string) (string, error)
	LaunchGroup(ctx context.Context, token string) (string, error)
	ExchangeLaunch(ctx context.Context, token, userID string) (string, error)
}
