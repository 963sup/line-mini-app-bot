package line

import (
	"regexp"
	"time"

	"github.com/line/line-bot-sdk-go/v8/linebot/webhook"
)

var groupID = regexp.MustCompile(`^C[0-9a-f]{32}$`)

func ValidGroupID(id string) bool { return groupID.MatchString(id) }

// GroupEvent contains protocol facts, not membership or domain authorization.
type GroupEvent struct {
	GroupID, EventID, Status string
	At                       time.Time
	Users                    []string
	Joined                   bool
}

// DecodeGroupEvent is called only after the webhook transport verifies signature.
func DecodeGroupEvent(event webhook.EventInterface) (GroupEvent, bool) {
	result := GroupEvent{Status: "ACTIVE"}
	var source webhook.SourceInterface
	var at int64
	var mode webhook.EventMode
	switch e := event.(type) {
	case webhook.JoinEvent:
		source, at, result.EventID, mode = e.Source, e.Timestamp, e.WebhookEventId, e.Mode
		result.Joined = true
	case webhook.LeaveEvent:
		source, at, result.EventID, mode = e.Source, e.Timestamp, e.WebhookEventId, e.Mode
		result.Status = "INACTIVE"
	case webhook.MessageEvent:
		source, at, result.EventID, mode = e.Source, e.Timestamp, e.WebhookEventId, e.Mode
	case webhook.PostbackEvent:
		source, at, result.EventID, mode = e.Source, e.Timestamp, e.WebhookEventId, e.Mode
	case webhook.MemberJoinedEvent:
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
	if !ok || mode != "active" || at <= 0 || result.EventID == "" || !ValidGroupID(g.GroupId) {
		return GroupEvent{}, false
	}
	result.GroupID = g.GroupId
	result.At = time.UnixMilli(at).UTC()
	result.Users = append(result.Users, g.UserId)
	return result, true
}
