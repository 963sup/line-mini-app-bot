package line

import (
	repositoryApplicationPort "github.com/963sup/line-mini-app-bot/internal/business/repository/application/port"
	repositoryDomainEntity "github.com/963sup/line-mini-app-bot/internal/business/repository/domain/entity"
)

import (
	"context"
	"errors"
	"time"

	"github.com/963sup/line-mini-app-bot/internal/platform/line/lineid"

	transport "github.com/963sup/line-mini-app-bot/internal/transport/line"
	"github.com/line/line-bot-sdk-go/v8/linebot/webhook"
)

type Observer struct {
	Store        repositoryApplicationPort.Store
	GroupName    func(context.Context, string) string
	ResolveOwner func(context.Context, string) (string, string, error)
}

// Observe uses the group facts parsed by the webhook dispatcher.
func (o *Observer) Observe(ctx context.Context, event webhook.EventInterface, group *transport.GroupEvent) error {
	if o == nil || o.Store == nil {
		return nil
	}
	if follow, ok := event.(webhook.FollowEvent); ok && follow.Mode == "active" && follow.Timestamp > 0 && follow.WebhookEventId != "" {
		if person, ok := follow.Source.(webhook.UserSource); ok && lineid.ValidUser(person.UserId) {
			err := o.Store.Observe(ctx, repositoryDomainEntity.Observation{OwnerType: repositoryDomainEntity.OwnerUser, OwnerID: person.UserId, Name: "個人儲存庫", Status: "ACTIVE", EventID: follow.WebhookEventId, At: time.UnixMilli(follow.Timestamp), Users: []string{person.UserId}})
			return err
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
			current, err := o.Store.GetByGroupID(ctx, group.GroupID)
			if err != nil && !errors.Is(err, repositoryDomainEntity.ErrNotFound) {
				return err
			}
			lookupName = errors.Is(err, repositoryDomainEntity.ErrNotFound) || current.Name == ""
		}
		if lookupName {
			name = o.GroupName(ctx, group.GroupID)
		}
	}
	owner := group.ActorID
	ownerType := repositoryDomainEntity.OwnerUser
	if status == "ACTIVE" && o.ResolveOwner != nil {
		resolvedType, resolvedID, resolveErr := o.ResolveOwner(ctx, group.ActorID)
		if resolveErr != nil {
			return resolveErr
		}
		ownerType, owner = repositoryDomainEntity.OwnerType(resolvedType), resolvedID
	}
	if status == "INACTIVE" {
		current, lookupErr := o.Store.GetByGroupID(ctx, group.GroupID)
		if errors.Is(lookupErr, repositoryDomainEntity.ErrNotFound) {
			return nil
		}
		if lookupErr != nil {
			return lookupErr
		}
		owner = current.OwnerID
		ownerType = current.OwnerType
	}
	err := o.Store.Observe(ctx, repositoryDomainEntity.Observation{OwnerType: ownerType, OwnerID: owner, GroupID: group.GroupID, Name: name, Status: status, EventID: group.EventID, At: group.At, Users: group.Users})
	return err
}
