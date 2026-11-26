package navigation

import (
	"context"
	"errors"
	"testing"

	transport "github.com/963sup/line-mini-app-bot/internal/transport/line"

	"github.com/line/line-bot-sdk-go/v8/linebot/messaging_api"
	"github.com/line/line-bot-sdk-go/v8/linebot/webhook"
)

func (trigger Menu) ReplyForEvent(event webhook.EventInterface, appURL, endpointURL string) transport.Reply {
	return trigger.parseMenu(event, trigger.routes()).reply(appURL, endpointURL)
}

func TestMiniAppPostbackReturnsDualEntryButtonsTemplate(t *testing.T) {
	t.Setenv("LINE_MINI_APP_WEB_URL", "https://attacker.example/miniapp")
	e := webhook.PostbackEvent{Mode: "active", ReplyToken: "reply", Source: webhook.GroupSource{GroupId: "group"}, Postback: &webhook.PostbackContent{Data: "action=miniapp"}}
	reply := (Menu{}).ReplyForEvent(e, "https://miniapp.line.me/123-test?route=projects", "https://app.example/miniapp?view=desktop")
	card, ok := reply.Message.(*messaging_api.TemplateMessage)
	if !ok || reply.Token != "reply" {
		t.Fatal("MINI App postback must return a buttons template")
	}
	action := card.Template.(*messaging_api.ButtonsTemplate).Actions[0].(*messaging_api.UriAction)
	if action.Uri != "https://miniapp.line.me/123-test?route=projects" || action.AltUri == nil || action.AltUri.Desktop != "https://app.example/miniapp?view=desktop" {
		t.Fatal("mobile and desktop destinations must remain separate")
	}
}

func TestMiniAppPostbackValidation(t *testing.T) {
	for _, data := range []string{"action=unknown", "action=miniapp&action=help", "action=miniapp&desktop=https://attacker.example"} {
		e := webhook.PostbackEvent{Mode: "active", ReplyToken: "reply", Source: webhook.UserSource{UserId: "user"}, Postback: &webhook.PostbackContent{Data: data}}
		if (Menu{}).ReplyForEvent(e, "https://miniapp.line.me/123-test", "https://app.example/miniapp").Message != nil {
			t.Fatal("only the exact supported postback is allowed")
		}
	}
	e := webhook.PostbackEvent{Mode: "standby", ReplyToken: "reply", Source: webhook.UserSource{UserId: "user"}, Postback: &webhook.PostbackContent{Data: "action=miniapp"}}
	if (Menu{}).ReplyForEvent(e, "https://miniapp.line.me/123-test", "https://app.example/miniapp").Message != nil {
		t.Fatal("standby event must remain silent")
	}
}

func TestGroupMenuLookup(t *testing.T) {
	event := webhook.PostbackEvent{Mode: "active", ReplyToken: "reply", Source: webhook.GroupSource{GroupId: "group"}, Postback: &webhook.PostbackContent{Data: "action=miniapp"}}
	unavailable := errors.New("repository unavailable")
	calls := 0
	handler := (Menu{}).Handler("https://example.org/miniapp", "", func(_ context.Context, group string) (string, error) {
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

func TestChatAndGroupMenu(t *testing.T) {
	const app = "https://miniapp.line.me/example"
	const repositoryID = "0123456789abcdef0123456789abcdef"
	lookups := 0
	handler := (Menu{}).Handler(app, "", func(_ context.Context, group string) (string, error) {
		lookups++
		if group != "group" {
			t.Fatalf("unexpected group %q", group)
		}
		return repositoryID, nil
	})
	user := webhook.UserSource{UserId: "user"}
	group := webhook.GroupSource{GroupId: "group", UserId: "user"}
	for _, source := range []webhook.SourceInterface{user, group} {
		e := webhook.MessageEvent{Source: source, Mode: "active", ReplyToken: "reply", Message: webhook.TextMessageContent{Text: "我剛剛忘了打卡"}}
		if reply, handled, err := handler.Handle(t.Context(), e); err != nil || handled || reply.Message != nil {
			t.Fatalf("general chat must remain quiet: %+v, %v", reply, err)
		}
		e.Message = webhook.TextMessageContent{Text: "功能"}
		reply, handled, err := handler.Handle(t.Context(), e)
		message, ok := reply.Message.(*messaging_api.TextMessage)
		if !handled || err != nil || !ok {
			t.Fatalf("explicit help must reply: %+v, %v", reply, err)
		}
		switch source.(type) {
		case webhook.UserSource:
			if message.QuickReply != nil || lookups != 0 {
				t.Fatal("private help should point to Rich Menu")
			}
		case webhook.GroupSource:
			if message.QuickReply == nil || len(message.QuickReply.Items) != 2 || lookups != 1 {
				t.Fatal("group help should expose two group links")
			}
			for i, want := range []string{app + "?group=group&menu=attendance", app + "?repositories=1&repository=" + repositoryID} {
				if got := message.QuickReply.Items[i].Action.(*messaging_api.UriAction).Uri; got != want {
					t.Fatalf("group link = %q, want %q", got, want)
				}
			}
		}
	}
	event := webhook.JoinEvent{Source: group, Mode: "active", ReplyToken: "reply"}
	if reply, handled, err := handler.Handle(t.Context(), event); err != nil || handled || reply.Message != nil || lookups != 1 {
		t.Fatalf("group join must remain quiet: %+v, %v", reply, err)
	}
}
