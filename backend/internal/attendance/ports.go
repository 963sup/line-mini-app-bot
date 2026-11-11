package attendance

import (
	"context"
	"time"
)

// PunchSubmitter records a located action bound to the authenticated user.
type PunchSubmitter interface {
	SubmitPunch(ctx context.Context, requestID, userID string, position Position, now time.Time) (Result, error)
}

// GroupHistoryReader reads only the authenticated user's group history.
type GroupHistoryReader interface {
	Groups(ctx context.Context, userID string) ([]string, error)
	RecentGroup(ctx context.Context, userID, groupID string) ([]Record, error)
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

type Store interface {
	Punch(ctx context.Context, groupID, userID, eventID, action string, now time.Time, eventTimeMillis int64) (Result, error)
	Recent(ctx context.Context, userID string) ([]Record, error)
}
