package line

import (
	"time"

	lineplatform "github.com/963sup/line-mini-app-bot/internal/platform/line"

	"github.com/line/line-bot-sdk-go/v8/linebot/webhook"
)

type GroupEventKind string

const (
	GroupBotJoined    GroupEventKind = "join"
	GroupBotLeft      GroupEventKind = "leave"
	GroupMessage      GroupEventKind = "message"
	GroupPostback     GroupEventKind = "postback"
	GroupMemberJoined GroupEventKind = "memberJoined"
)

// GroupEvent contains LINE protocol facts. Consumers decide their business state.
type GroupEvent struct {
	GroupID, EventID string
	Kind             GroupEventKind
	At               time.Time
	Users            []string
}

// DecodeGroupEvent is called only after the webhook transport verifies signature.
func DecodeGroupEvent(event webhook.EventInterface) (GroupEvent, bool) {
	var result GroupEvent
	var source webhook.SourceInterface
	var at int64
	var mode webhook.EventMode
	switch e := event.(type) {
	case webhook.JoinEvent:
		source, at, result.EventID, mode = e.Source, e.Timestamp, e.WebhookEventId, e.Mode
		result.Kind = GroupBotJoined
	case webhook.LeaveEvent:
		source, at, result.EventID, mode = e.Source, e.Timestamp, e.WebhookEventId, e.Mode
		result.Kind = GroupBotLeft
	case webhook.MessageEvent:
		result.Kind = GroupMessage
		source, at, result.EventID, mode = e.Source, e.Timestamp, e.WebhookEventId, e.Mode
	case webhook.PostbackEvent:
		result.Kind = GroupPostback
		source, at, result.EventID, mode = e.Source, e.Timestamp, e.WebhookEventId, e.Mode
	case webhook.MemberJoinedEvent:
		result.Kind = GroupMemberJoined
		source, at, result.EventID, mode = e.Source, e.Timestamp, e.WebhookEventId, e.Mode
		if e.Joined != nil {
			for _, member := range e.Joined.Members {
				result.Users = append(result.Users, member.UserId)
			}
		}
	default:
		return GroupEvent{}, false
	}
	g, ok := source.(webhook.GroupSource)
	if !ok || mode != "active" || at <= 0 || result.EventID == "" || !lineplatform.ValidGroupID(g.GroupId) {
		return GroupEvent{}, false
	}
	result.GroupID = g.GroupId
	result.At = time.UnixMilli(at).UTC()
	result.Users = append(result.Users, g.UserId)
	return result, true
}
