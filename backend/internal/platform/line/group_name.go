package line

import (
	"context"
	"time"

	"github.com/line/line-bot-sdk-go/v8/linebot/messaging_api"
)

// GroupNameLookup returns a best-effort lookup for the LINE group display name.
func GroupNameLookup(bot *messaging_api.MessagingApiAPI) func(context.Context, string) string {
	return func(ctx context.Context, group string) string {
		lookupCtx, cancel := context.WithTimeout(ctx, 400*time.Millisecond)
		defer cancel()
		client := *bot
		client.WithContext(lookupCtx)
		summary, err := client.GetGroupSummary(group)
		if err != nil || summary == nil {
			return ""
		}
		return summary.GroupName
	}
}
