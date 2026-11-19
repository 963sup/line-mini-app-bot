package line

import (
	"testing"

	"github.com/line/line-bot-sdk-go/v8/linebot/messaging_api"
	"github.com/line/line-bot-sdk-go/v8/linebot/webhook"
)

func TestMiniAppPostbackReturnsDualEntryFlex(t *testing.T) {
	t.Setenv("LINE_MINI_APP_WEB_URL", "https://attacker.example/miniapp")
	e := webhook.PostbackEvent{Mode: "active", ReplyToken: "reply", Source: webhook.GroupSource{GroupId: "group"}, Postback: &webhook.PostbackContent{Data: "action=miniapp"}}
	reply := ReplyForEvent(e, "https://miniapp.line.me/123-test?route=projects")
	card, ok := reply.Message.(*messaging_api.FlexMessage)
	if !ok || reply.Token != "reply" {
		t.Fatal("MINI App postback must return a Flex card")
	}
	bubble := card.Contents.(*messaging_api.FlexBubble)
	action := bubble.Footer.Contents[0].(*messaging_api.FlexButton).Action.(*messaging_api.UriAction)
	if action.Uri != "https://miniapp.line.me/123-test?route=projects" || action.AltUri == nil || action.AltUri.Desktop != "https://line-mini-app-bot-jtghegqd7q-de.a.run.app/miniapp" {
		t.Fatal("mobile and desktop destinations must remain separate")
	}
}

func TestMiniAppPostbackValidation(t *testing.T) {
	for _, data := range []string{"action=unknown", "action=miniapp&action=help", "action=miniapp&desktop=https://attacker.example"} {
		e := webhook.PostbackEvent{Mode: "active", ReplyToken: "reply", Source: webhook.UserSource{UserId: "user"}, Postback: &webhook.PostbackContent{Data: data}}
		if ReplyForEvent(e, "https://miniapp.line.me/123-test").Message != nil {
			t.Fatal("only the exact supported postback is allowed")
		}
	}
	e := webhook.PostbackEvent{Mode: "standby", ReplyToken: "reply", Source: webhook.UserSource{UserId: "user"}, Postback: &webhook.PostbackContent{Data: "action=miniapp"}}
	if ReplyForEvent(e, "https://miniapp.line.me/123-test").Message != nil {
		t.Fatal("standby event must remain silent")
	}
}
