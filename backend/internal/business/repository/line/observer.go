package line

import (
	"context"
	"errors"
	"time"

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
	var source webhook.SourceInterface
	var at int64
	var id string
	var mode webhook.EventMode
	state := "ACTIVE"
	var users []string
	join := false
	switch e := event.(type) {
	case webhook.JoinEvent:
		source, at, id, mode = e.Source, e.Timestamp, e.WebhookEventId, e.Mode
		join = true
	case webhook.LeaveEvent:
		source, at, id, mode = e.Source, e.Timestamp, e.WebhookEventId, e.Mode
		state = "INACTIVE"
	case webhook.MessageEvent:
		source, at, id, mode = e.Source, e.Timestamp, e.WebhookEventId, e.Mode
	case webhook.PostbackEvent:
		source, at, id, mode = e.Source, e.Timestamp, e.WebhookEventId, e.Mode
	case webhook.MemberJoinedEvent:
		source, at, id, mode = e.Source, e.Timestamp, e.WebhookEventId, e.Mode
		if e.Joined != nil {
			for _, member := range e.Joined.Members {
				users = append(users, member.UserId)
			}
		}
	default:
		return transport.Reply{}, false, nil
	}
	g, ok := source.(webhook.GroupSource)
	if !ok || mode != "active" || at <= 0 || id == "" || !repository.ValidGroup(g.GroupId) {
		return transport.Reply{}, false, nil
	}
	users = append(users, g.UserId)
	name := ""
	if state == "ACTIVE" && o.GroupName != nil {
		lookupName := join
		if !lookupName {
			current, err := o.Store.GetByGroupID(ctx, g.GroupId)
			if err != nil && !errors.Is(err, repository.ErrNotFound) {
				return transport.Reply{}, false, err
			}
			lookupName = errors.Is(err, repository.ErrNotFound) || current.Name == ""
		}
		if lookupName {
			name = o.GroupName(ctx, g.GroupId)
		}
	}
	err := o.Store.Observe(ctx, repository.Observation{GroupID: g.GroupId, Name: name, Status: state, EventID: id, At: time.UnixMilli(at).UTC(), Users: users})
	return transport.Reply{}, false, err
}
