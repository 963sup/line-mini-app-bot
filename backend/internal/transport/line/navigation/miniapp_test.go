package navigation

import (
	"context"
	"errors"
	"testing"

	"github.com/line/line-bot-sdk-go/v8/linebot/messaging_api"
	"github.com/line/line-bot-sdk-go/v8/linebot/webhook"
)

func TestMiniAppPostbackReturnsDualEntryFlex(t *testing.T) {
	t.Setenv("LINE_MINI_APP_WEB_URL", "https://attacker.example/miniapp")
	e := webhook.PostbackEvent{Mode: "active", ReplyToken: "reply", Source: webhook.GroupSource{GroupId: "group"}, Postback: &webhook.PostbackContent{Data: "action=miniapp"}}
	reply := ReplyForEvent(e, "https://miniapp.line.me/123-test?route=projects", "https://app.example/miniapp?view=desktop")
	card, ok := reply.Message.(*messaging_api.FlexMessage)
	if !ok || reply.Token != "reply" {
		t.Fatal("MINI App postback must return a Flex card")
	}
	bubble := card.Contents.(*messaging_api.FlexBubble)
	action := bubble.Footer.Contents[0].(*messaging_api.FlexButton).Action.(*messaging_api.UriAction)
	if action.Uri != "https://miniapp.line.me/123-test?route=projects" || action.AltUri == nil || action.AltUri.Desktop != "https://app.example/miniapp?view=desktop" {
		t.Fatal("mobile and desktop destinations must remain separate")
	}
}

func TestMiniAppPostbackValidation(t *testing.T) {
	for _, data := range []string{"action=unknown", "action=miniapp&action=help", "action=miniapp&desktop=https://attacker.example"} {
		e := webhook.PostbackEvent{Mode: "active", ReplyToken: "reply", Source: webhook.UserSource{UserId: "user"}, Postback: &webhook.PostbackContent{Data: data}}
		if ReplyForEvent(e, "https://miniapp.line.me/123-test", "https://app.example/miniapp").Message != nil {
			t.Fatal("only the exact supported postback is allowed")
		}
	}
	e := webhook.PostbackEvent{Mode: "standby", ReplyToken: "reply", Source: webhook.UserSource{UserId: "user"}, Postback: &webhook.PostbackContent{Data: "action=miniapp"}}
	if ReplyForEvent(e, "https://miniapp.line.me/123-test", "https://app.example/miniapp").Message != nil {
		t.Fatal("standby event must remain silent")
	}
}

func TestGroupMenuLookup(t *testing.T) {
	event := webhook.PostbackEvent{Mode: "active", ReplyToken: "reply", Source: webhook.GroupSource{GroupId: "group"}, Postback: &webhook.PostbackContent{Data: "action=miniapp"}}
	unavailable := errors.New("repository unavailable")
	calls := 0
	handler := Handler("https://example.org/miniapp", "", func(_ context.Context, group string) (string, error) {
		calls++
		if group != "group" {
			t.Fatalf("unexpected group %q", group)
		}
		return "", unavailable
	})
	reply, handled, err := handler.Handle(t.Context(), event)
	if !errors.Is(err, unavailable) || handled || reply.Message != nil || calls != 1 {
		t.Fatal("group lookup failure must be returned before replying")
	}
	event.Postback.Data = "action=unknown"
	reply, _, err = handler.Handle(t.Context(), event)
	if err != nil || reply.Message != nil || calls != 1 {
		t.Fatal("ignored postback must not query repository")
	}
}
