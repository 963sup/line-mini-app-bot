package line

import (
	"testing"

	"github.com/line/line-bot-sdk-go/v8/linebot/messaging_api"
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

func TestMenuActionsKeepDisplayLabelsSeparateFromCommands(t *testing.T) {
	actions := MenuActions(true, "", true)
	wantLabels := []string{"打卡", "出勤", "請假登記", "補卡申請"}
	wantCommands := []string{"打卡", "出勤", "請假", "補卡"}
	for index, action := range actions {
		message, ok := action.(*messaging_api.MessageAction)
		if !ok || message.Label != wantLabels[index] || message.Text != wantCommands[index] {
			t.Fatalf("action %d: %#v", index, action)
		}
	}
}
