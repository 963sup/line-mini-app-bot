package line

import (
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/line/line-bot-sdk-go/v8/linebot/webhook"
)

func TestDecodeGroupEvent(t *testing.T) {
	source := webhook.GroupSource{GroupId: "C" + strings.Repeat("a", 32), UserId: "actor"}
	const at int64 = 1700000000123
	cases := []struct {
		name   string
		event  webhook.EventInterface
		state  string
		joined bool
		users  []string
	}{
		{"join", webhook.JoinEvent{Source: source, Timestamp: at, Mode: "active", WebhookEventId: "event"}, "ACTIVE", true, []string{"actor"}},
		{"join without user", webhook.JoinEvent{Source: webhook.GroupSource{GroupId: source.GroupId}, Timestamp: at, Mode: "active", WebhookEventId: "event"}, "ACTIVE", true, []string{""}},
		{"leave", webhook.LeaveEvent{Source: source, Timestamp: at, Mode: "active", WebhookEventId: "event"}, "INACTIVE", false, []string{"actor"}},
		{"leave without user", webhook.LeaveEvent{Source: webhook.GroupSource{GroupId: source.GroupId}, Timestamp: at, Mode: "active", WebhookEventId: "event"}, "INACTIVE", false, []string{""}},
		{"message", webhook.MessageEvent{Source: source, Timestamp: at, Mode: "active", WebhookEventId: "event"}, "ACTIVE", false, []string{"actor"}},
		{"postback", webhook.PostbackEvent{Source: source, Timestamp: at, Mode: "active", WebhookEventId: "event"}, "ACTIVE", false, []string{"actor"}},
		{"members", webhook.MemberJoinedEvent{Source: source, Timestamp: at, Mode: "active", WebhookEventId: "event", Joined: &webhook.JoinedMembers{Members: []webhook.UserSource{{UserId: "member"}, {UserId: "actor"}}}}, "ACTIVE", false, []string{"member", "actor", "actor"}},
		{"absent members", webhook.MemberJoinedEvent{Source: source, Timestamp: at, Mode: "active", WebhookEventId: "event"}, "ACTIVE", false, []string{"actor"}},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got, ok := DecodeGroupEvent(tc.event)
			if !ok || got.GroupID != source.GroupId || got.EventID != "event" || got.Status != tc.state || got.Joined != tc.joined || !got.At.Equal(time.UnixMilli(at)) || got.At.Location() != time.UTC || !reflect.DeepEqual(got.Users, tc.users) {
				t.Fatalf("unexpected protocol facts: %+v, accepted=%v", got, ok)
			}
		})
	}
}

func TestDecodeGroupEventRejectsInvalidProtocolFacts(t *testing.T) {
	valid := webhook.JoinEvent{Source: webhook.GroupSource{GroupId: "C" + strings.Repeat("a", 32)}, Timestamp: 1, Mode: "active", WebhookEventId: "event"}
	cases := map[string]webhook.EventInterface{"unsupported": webhook.UnfollowEvent{}, "nil": nil}
	for _, name := range []string{"standby", "missing mode", "zero time", "negative time", "missing ID", "wrong group", "uppercase group", "short group", "user source"} {
		event := valid
		switch name {
		case "standby":
			event.Mode = "standby"
		case "missing mode":
			event.Mode = ""
		case "zero time":
			event.Timestamp = 0
		case "negative time":
			event.Timestamp = -1
		case "missing ID":
			event.WebhookEventId = ""
		case "wrong group":
			event.Source = webhook.GroupSource{GroupId: "U" + strings.Repeat("a", 32)}
		case "uppercase group":
			event.Source = webhook.GroupSource{GroupId: "C" + strings.Repeat("A", 32)}
		case "short group":
			event.Source = webhook.GroupSource{GroupId: "Ca"}
		case "user source":
			event.Source = webhook.UserSource{UserId: "actor"}
		}
		cases[name] = event
	}
	for name, event := range cases {
		t.Run(name, func(t *testing.T) {
			if got, ok := DecodeGroupEvent(event); ok || !reflect.DeepEqual(got, GroupEvent{}) {
				t.Fatalf("invalid event accepted: %+v", got)
			}
		})
	}
}
