package line

import (
	"context"
	"errors"

	groupentity "github.com/963sup/line-mini-app-bot/internal/linegroup/domain/entity"
	groupevent "github.com/963sup/line-mini-app-bot/internal/linegroup/domain/event"
	transport "github.com/963sup/line-mini-app-bot/internal/transport/line"
	"github.com/line/line-bot-sdk-go/v8/linebot/webhook"
)

type Observer struct {
	Store interface {
		Observe(context.Context, groupevent.Observation) error
		GetByID(context.Context, string) (groupentity.Group, error)
	}
	GroupName func(context.Context, string) string
}

// Observe uses the group facts parsed by the webhook dispatcher.
func (o *Observer) Observe(ctx context.Context, _ webhook.EventInterface, group *transport.GroupEvent) error {
	if o == nil || o.Store == nil {
		return nil
	}
	if group == nil {
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
			current, err := o.Store.GetByID(ctx, group.GroupID)
			if err != nil && !errors.Is(err, groupentity.ErrNotFound) {
				return err
			}
			lookupName = errors.Is(err, groupentity.ErrNotFound) || current.Name == ""
		}
		if lookupName {
			name = o.GroupName(ctx, group.GroupID)
		}
	}
	err := o.Store.Observe(ctx, groupevent.Observation{GroupID: group.GroupID, Name: name, Status: status, EventID: group.EventID, At: group.At, Users: group.Users})
	return err
}
