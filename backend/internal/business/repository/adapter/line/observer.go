package line

import (
	repositoryApplication "github.com/963sup/line-mini-app-bot/internal/business/repository/application"
)

import (
	"context"
	"time"

	"github.com/963sup/line-mini-app-bot/internal/platform/line/lineid"

	transport "github.com/963sup/line-mini-app-bot/internal/transport/line"
	"github.com/line/line-bot-sdk-go/v8/linebot/webhook"
)

type Observer struct {
	Bindings  *repositoryApplication.Bindings
	GroupName func(context.Context, string) string
}

// Observe uses the group facts parsed by the webhook dispatcher.
func (o *Observer) Observe(ctx context.Context, event webhook.EventInterface, group *transport.GroupEvent) error {
	if o == nil || o.Bindings == nil {
		return nil
	}
	if follow, ok := event.(webhook.FollowEvent); ok && follow.Mode == "active" && follow.Timestamp > 0 && follow.WebhookEventId != "" {
		if person, ok := follow.Source.(webhook.UserSource); ok && lineid.ValidUser(person.UserId) {
			return o.Bindings.ObservePersonal(ctx, person.UserId, "個人儲存庫", follow.WebhookEventId, time.UnixMilli(follow.Timestamp))
		}
	}
	if group == nil {
		return nil
	}
	if !lineid.ValidUser(group.ActorID) && group.Kind != transport.GroupBotLeft {
		return nil
	}
	// Preserve the existing lifecycle: supported group activity records an
	// active observation, while the bot leaving records an inactive one.
	var status string
	switch group.Kind {
	case transport.GroupBotLeft:
		status = "INACTIVE"
	case transport.GroupBotJoined, transport.GroupMessage, transport.GroupPostback, transport.GroupMemberJoined:
		status = "ACTIVE"
	default:
		return nil
	}
	name := ""
	if status == "ACTIVE" && o.GroupName != nil {
		lookupName := group.Kind == transport.GroupBotJoined
		if !lookupName {
			var err error
			lookupName, err = o.Bindings.NeedsName(ctx, group.GroupID)
			if err != nil {
				return err
			}
		}
		if lookupName {
			name = o.GroupName(ctx, group.GroupID)
		}
	}
	return o.Bindings.ObserveGroup(ctx, group.ActorID, group.GroupID, name, status, group.EventID, group.At, group.Users)
}
