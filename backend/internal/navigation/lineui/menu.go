// Package lineui renders the public feature menu shared by the bot and CLI.
package lineui

import (
	"net/url"
	"regexp"

	"github.com/line/line-bot-sdk-go/v8/linebot/messaging_api"
)

func journalURL(appURL string) string {
	if appURL == "" {
		return ""
	}
	u, err := url.Parse(appURL)
	if err != nil || u.Scheme != "https" || u.Host == "" {
		return ""
	}
	q := u.Query()
	q.Set("journal", "1")
	u.RawQuery = q.Encode()
	return u.String()
}

func repositoryURL(appURL string, group bool, ids ...string) string {
	if group && (len(ids) != 1 || !regexp.MustCompile(`^[a-f0-9]{32}$`).MatchString(ids[0])) {
		return ""
	}
	if appURL == "" {
		return ""
	}
	u, err := url.Parse(appURL)
	if err != nil || u.Scheme != "https" || u.Host == "" {
		return ""
	}
	q := u.Query()
	q.Set("repositories", "1")
	if group {
		q.Set("repository", ids[0])
	}
	u.RawQuery = q.Encode()
	return u.String()
}

// AttendanceRichMenus previews two tabs. Aliases must be bound to uploaded menu
// IDs before publishing. Switching tabs never records attendance.
func AttendanceRichMenus(appURL string) map[string]*messaging_api.RichMenuRequest {
	menus := make(map[string]*messaging_api.RichMenuRequest)
	for _, tab := range []struct{ alias, other, label, action string }{
		{"attendance-in", "attendance-out", "出勤紀錄", "attendance-history"},
		{"attendance-out", "attendance-in", "功能說明", "help"},
	} {
		menu := RichMenu(appURL)
		menu.Name = "出勤：" + tab.label
		menu.Areas[0].Action = &messaging_api.PostbackAction{Label: tab.label, Data: "action=" + tab.action}
		menu.Areas[1].Action = &messaging_api.RichMenuSwitchAction{Label: "切換分頁", RichMenuAliasId: tab.other, Data: "tab=" + tab.other}
		menus[tab.alias] = menu
	}
	return menus
}

func Help(group bool, appURL string, repositoryIDs ...string) messaging_api.MessageInterface {
	title := "LINE 功能選單"
	description := "輸入「出勤紀錄」查詢本人紀錄，或開啟網頁服務。"
	if group {
		title = "群組功能選單"
		description = "點選下方按鈕：打卡、出勤、請假登記或補卡申請。請假送出不代表核准；歷史出勤請私訊查詢。分享位置可查天氣，輸入「摘要」整理群聊。"
	}
	buttons := []messaging_api.FlexComponentInterface{}
	if group {
		for _, entry := range []struct{ label, command string }{
			{"打卡", "打卡"}, {"出勤", "出勤"}, {"請假登記", "請假"}, {"補卡申請", "補卡"},
		} {
			buttons = append(buttons, &messaging_api.FlexButton{Action: &messaging_api.MessageAction{Label: entry.label, Text: entry.command}})
			if entry.command == "打卡" && journalURL(appURL) != "" {
				buttons = append(buttons, &messaging_api.FlexButton{Action: &messaging_api.UriAction{Label: "日誌", Uri: journalURL(appURL)}})
			}
		}
	} else {
		buttons = append(buttons, &messaging_api.FlexButton{Action: &messaging_api.PostbackAction{Label: "出勤紀錄", Data: "action=attendance-history"}})
	}
	if link := repositoryURL(appURL, group, repositoryIDs...); link != "" {
		buttons = append(buttons, &messaging_api.FlexButton{Action: &messaging_api.UriAction{Label: "空間", Uri: link}})
	}
	if appURL != "" {
		buttons = append(buttons, &messaging_api.FlexButton{
			Style:  "primary",
			Action: &messaging_api.UriAction{Label: "開啟 MINI App", Uri: appURL},
		})
	}
	return &messaging_api.FlexMessage{
		AltText: title,
		Contents: &messaging_api.FlexBubble{
			Body: &messaging_api.FlexBox{
				Layout: "vertical",
				Contents: []messaging_api.FlexComponentInterface{
					&messaging_api.FlexText{Text: title, Weight: "bold", Size: "xl", Wrap: true},
					&messaging_api.FlexText{Text: description, Wrap: true, Margin: "md"},
				},
			},
			Footer: &messaging_api.FlexBox{Layout: "vertical", Contents: buttons},
		},
	}
}

// QuickMenu puts each supported command in a separate reply chip.
func QuickMenu(group bool, appURL string, repositoryIDs ...string) messaging_api.MessageInterface {
	items := []messaging_api.QuickReplyItem{}
	add := func(label, command string) {
		items = append(items, messaging_api.QuickReplyItem{Type: "action", Action: &messaging_api.MessageAction{Label: label, Text: command}})
	}
	if group {
		add("打卡", "打卡")
		if link := journalURL(appURL); link != "" {
			items = append(items, messaging_api.QuickReplyItem{Type: "action", Action: &messaging_api.UriAction{Label: "日誌", Uri: link}})
		}
		add("出勤", "出勤")
		add("請假", "請假")
		add("補卡", "補卡")
		add("摘要", "摘要")
	} else {
		add("出勤紀錄", "出勤紀錄")
	}
	if link := repositoryURL(appURL, group, repositoryIDs...); link != "" {
		items = append(items, messaging_api.QuickReplyItem{Type: "action", Action: &messaging_api.UriAction{Label: "空間", Uri: link}})
	}
	if appURL != "" {
		items = append(items, messaging_api.QuickReplyItem{Type: "action", Action: &messaging_api.UriAction{Label: "MINI App", Uri: appURL}})
	}
	return &messaging_api.TextMessage{Text: "請選擇功能：", QuickReply: &messaging_api.QuickReply{Items: items}}
}

// RichMenu describes three horizontal areas matching a 2500 x 843 image.
func RichMenu(appURL string) *messaging_api.RichMenuRequest {
	return &messaging_api.RichMenuRequest{
		Size: &messaging_api.RichMenuSize{Width: 2500, Height: 843},
		Name: "LINE 基礎服務", ChatBarText: "開啟選單", Selected: false,
		Areas: []messaging_api.RichMenuArea{
			{
				Bounds: &messaging_api.RichMenuBounds{X: 0, Y: 0, Width: 833, Height: 843},
				Action: &messaging_api.PostbackAction{Label: "功能說明", Data: "action=help"},
			},
			{
				Bounds: &messaging_api.RichMenuBounds{X: 833, Y: 0, Width: 833, Height: 843},
				Action: &messaging_api.PostbackAction{Label: "出勤紀錄", Data: "action=attendance-history"},
			},
			{
				Bounds: &messaging_api.RichMenuBounds{X: 1666, Y: 0, Width: 834, Height: 843},
				Action: &messaging_api.UriAction{Label: "開啟 MINI App", Uri: appURL},
			},
		},
	}
}
