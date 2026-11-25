package line

import (
	"context"
	"testing"

	"github.com/line/line-bot-sdk-go/v8/linebot/messaging_api"
	"github.com/line/line-bot-sdk-go/v8/linebot/webhook"
)

func TestAttendanceMenuUsesQuickReplyForTemporaryChoices(t *testing.T) {
	links := []string{"https://example.org/in", "https://example.org/out", "https://example.org/break", "https://example.org/resume", "https://example.org/confirm"}
	message, ok := AttendanceMenu(links).(*messaging_api.TextMessage)
	if !ok || message.QuickReply == nil || len(message.QuickReply.Items) != len(links) {
		t.Fatalf("attendance menu must be a five-item quick reply: %#v", message)
	}
	for index, item := range message.QuickReply.Items {
		action, ok := item.Action.(*messaging_api.UriAction)
		if !ok || action.Uri != links[index] {
			t.Fatalf("choice %d: %#v", index, item.Action)
		}
	}
}

func TestCorrectionEntriesKeepMissingPunchAndTimeCorrectionDistinct(t *testing.T) {
	bot := &Attendance{AppURL: "https://liff.line.me/example"}
	for _, tc := range []struct {
		command, label, uri string
	}{
		{"補卡", "申請補卡", "https://liff.line.me/example?correction=group"},
		{"更正", "選擇出勤紀錄", "https://liff.line.me/example?attendance=1"},
	} {
		event := webhook.MessageEvent{Mode: "active", ReplyToken: "reply", Source: webhook.GroupSource{GroupId: "group", UserId: "actor"}, Message: webhook.TextMessageContent{Text: tc.command}}
		reply, handled, err := bot.Reply(context.Background(), event)
		message, ok := reply.Message.(*messaging_api.TemplateMessage)
		if err != nil || !handled || !ok {
			t.Fatalf("%s: reply=%#v handled=%v err=%v", tc.command, reply, handled, err)
		}
		template := message.Template.(*messaging_api.ButtonsTemplate)
		action := template.Actions[0].(*messaging_api.UriAction)
		if action.Label != tc.label || action.Uri != tc.uri {
			t.Fatalf("%s: action=%#v", tc.command, action)
		}
	}
}

func TestMenuActionsKeepDisplayLabelsSeparateFromCommands(t *testing.T) {
	actions := MenuActions(true, "", true)
	wantLabels := []string{"打卡", "出勤", "請假", "補卡", "更正"}
	wantCommands := []string{"打卡", "出勤", "請假", "補卡", "更正"}
	for index, action := range actions {
		message, ok := action.(*messaging_api.MessageAction)
		if !ok || message.Label != wantLabels[index] || message.Text != wantCommands[index] {
			t.Fatalf("action %d: %#v", index, action)
		}
	}
}
