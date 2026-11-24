package bootstrap

import (
	attendance "github.com/963sup/line-mini-app-bot/internal/business/attendance/adapter/line"
	organization "github.com/963sup/line-mini-app-bot/internal/business/organization/adapter/line"
	summary "github.com/963sup/line-mini-app-bot/internal/business/summary/adapter/line"
	"github.com/963sup/line-mini-app-bot/internal/transport/line/navigation"
	"github.com/line/line-bot-sdk-go/v8/linebot/messaging_api"
)

// quickMenu puts each supported command in a separate reply chip.
func quickMenu(group bool, appURL string, repositoryIDs ...string) messaging_api.MessageInterface {
	items := []messaging_api.QuickReplyItem{}
	for _, action := range attendance.MenuActions(group, navigation.JournalURL(appURL, repositoryIDs...), true) {
		items = append(items, messaging_api.QuickReplyItem{Type: "action", Action: action})
	}
	commands := []string{organization.CreateOrganizationCommand, organization.CreateEnterpriseCommand}
	if group {
		commands = []string{summary.SummaryCommand}
	}
	for _, command := range commands {
		label := command
		if command == summary.SummaryCommand {
			label = "群聊摘要"
		}
		items = append(items, messaging_api.QuickReplyItem{Type: "action", Action: &messaging_api.MessageAction{Label: label, Text: command}})
	}
	if link := navigation.RepositoryURL(appURL, group, repositoryIDs...); link != "" {
		items = append(items, messaging_api.QuickReplyItem{Type: "action", Action: &messaging_api.UriAction{Label: "空間", Uri: link}})
	}
	if appURL != "" {
		items = append(items, messaging_api.QuickReplyItem{Type: "action", Action: &messaging_api.PostbackAction{Label: "MINI App", Data: "action=miniapp"}})
	}
	return &messaging_api.TextMessage{Text: "請選擇功能：", QuickReply: &messaging_api.QuickReply{Items: items}}
}
