package records

import "context"

// GroupHistoryReader reads only the authenticated user's group history.
type GroupHistoryReader interface {
	Groups(ctx context.Context, userID string) ([]string, error)
	RecentGroup(ctx context.Context, userID, groupID string) ([]Record, error)
}
