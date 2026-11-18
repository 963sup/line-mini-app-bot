package line

import (
	"context"
	"errors"
	"time"

	"github.com/963sup/line-mini-app-bot/internal/identity"

	"github.com/963sup/line-mini-app-bot/internal/business/repository"
	transport "github.com/963sup/line-mini-app-bot/internal/transport/line"
	"github.com/line/line-bot-sdk-go/v8/linebot/webhook"
)

type Observer struct {
	Store     repository.Store
	GroupName func(context.Context, string) string
}

// Observe never consumes the event: summary and attendance keep their handlers.
func (o *Observer) Observe(ctx context.Context, event webhook.EventInterface) (transport.Reply, bool, error) {
	if o == nil || o.Store == nil {
		return transport.Reply{}, false, nil
	}
	if follow, ok := event.(webhook.FollowEvent); ok && follow.Mode == "active" && follow.Timestamp > 0 && follow.WebhookEventId != "" {
		if person, ok := follow.Source.(webhook.UserSource); ok && identity.IsUserOwnerID(person.UserId) {
			err := o.Store.Observe(ctx, repository.Observation{OwnerID: person.UserId, Name: "個人儲存庫", Status: "ACTIVE", EventID: follow.WebhookEventId, At: time.UnixMilli(follow.Timestamp), Users: []string{person.UserId}})
			return transport.Reply{}, false, err
		}
	}
	group, ok := transport.DecodeGroupEvent(event)
	if !ok {
		return transport.Reply{}, false, nil
	}
	name := ""
	if group.Status == "ACTIVE" && o.GroupName != nil {
		lookupName := group.Joined
		if !lookupName {
			current, err := o.Store.GetByGroupID(ctx, group.GroupID)
			if err != nil && !errors.Is(err, repository.ErrNotFound) {
				return transport.Reply{}, false, err
			}
			lookupName = errors.Is(err, repository.ErrNotFound) || current.Name == ""
		}
		if lookupName {
			name = o.GroupName(ctx, group.GroupID)
		}
	}
	err := o.Store.Observe(ctx, repository.Observation{OwnerID: group.GroupID, Name: name, Status: group.Status, EventID: group.EventID, At: group.At, Users: group.Users})
	return transport.Reply{}, false, err
}
