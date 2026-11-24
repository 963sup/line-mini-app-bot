package line

import (
	"context"
	"time"
)

// GroupName returns an empty display name when the lookup times out or fails.
func (m *Messaging) GroupName(ctx context.Context, group string) string {
	lookupCtx, cancel := context.WithTimeout(ctx, 400*time.Millisecond)
	defer cancel()
	summary, err := m.request(lookupCtx).GetGroupSummary(group)
	if err != nil || summary == nil {
		return ""
	}
	return summary.GroupName
}
