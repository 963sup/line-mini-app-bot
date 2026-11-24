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
		items = append(items, messaging_api.QuickReplyItem{Type: "action", Action: &messaging_api.MessageAction{Label: command, Text: command}})
	}
	if link := navigation.RepositoryURL(appURL, group, repositoryIDs...); link != "" {
		items = append(items, messaging_api.QuickReplyItem{Type: "action", Action: &messaging_api.UriAction{Label: "空間", Uri: link}})
	}
	if appURL != "" {
		items = append(items, messaging_api.QuickReplyItem{Type: "action", Action: &messaging_api.PostbackAction{Label: "MINI App", Data: "action=miniapp"}})
	}
	return &messaging_api.TextMessage{Text: "請選擇功能：", QuickReply: &messaging_api.QuickReply{Items: items}}
}

func helpMenu(group bool, appURL, desktopURL string, repositoryIDs ...string) messaging_api.MessageInterface {
	title := "LINE 功能選單"
	description := "輸入「出勤紀錄」查詢本人紀錄，或輸入「建立組織」「建立企業」開始建立。"
	if group {
		title = "群組功能選單"
		description = "點選下方按鈕：打卡、出勤、請假登記或補卡申請。請假送出不代表核准；歷史出勤請私訊查詢。分享位置可查天氣，輸入「摘要」整理群聊。"
	}
	var buttons []messaging_api.FlexComponentInterface
	for _, action := range attendance.MenuActions(group, navigation.JournalURL(appURL, repositoryIDs...), false) {
		buttons = append(buttons, &messaging_api.FlexButton{Action: action})
	}
	if link := navigation.RepositoryURL(appURL, group, repositoryIDs...); link != "" {
		buttons = append(buttons, &messaging_api.FlexButton{Action: &messaging_api.UriAction{Label: "空間", Uri: link}})
	}
	return navigation.HelpCard(title, description, buttons, appURL, desktopURL)
}
