// Package lineui renders the public connectivity menu shared by the bot and CLI.
package lineui

import "github.com/line/line-bot-sdk-go/v8/linebot/messaging_api"

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

func Help(group bool, appURL string) messaging_api.MessageInterface {
	title := "LINE 功能選單"
	description := "一對一連線已接通。輸入「出勤紀錄」查詢本人紀錄，或開啟網頁服務。"
	if group {
		title = "群組功能選單"
		description = "輸入「打卡」叫出上下班按鈕；分享位置查天氣。輸入「摘要」整理最近群聊，「關閉摘要／開啟摘要」切換收集。歷史出勤請私訊查詢。"
	}
	buttons := []messaging_api.FlexComponentInterface{
		&messaging_api.FlexButton{
			Action: &messaging_api.PostbackAction{Label: "連線狀態", Data: "action=status"},
		},
	}
	if !group {
		buttons = append(buttons, &messaging_api.FlexButton{Action: &messaging_api.PostbackAction{Label: "出勤紀錄", Data: "action=attendance-history"}})
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
func QuickMenu(group bool, appURL string) messaging_api.MessageInterface {
	items := []messaging_api.QuickReplyItem{}
	add := func(label, command string) {
		items = append(items, messaging_api.QuickReplyItem{Type: "action", Action: &messaging_api.MessageAction{Label: label, Text: command}})
	}
	if group {
		add("打卡", "打卡")
		add("摘要", "摘要")
		add("關閉摘要", "關閉摘要")
		add("開啟摘要", "開啟摘要")
	} else {
		add("出勤紀錄", "出勤紀錄")
	}
	add("連線狀態", "/status")
	if appURL != "" {
		items = append(items, messaging_api.QuickReplyItem{Type: "action", Action: &messaging_api.UriAction{Label: "開啟 MINI App", Uri: appURL}})
	}
	return &messaging_api.TextMessage{Text: "請選擇功能：", QuickReply: &messaging_api.QuickReply{Items: items}}
}

// AttendanceMenu uses silent postbacks so a click doesn't send an extra text message.
func AttendanceMenu() messaging_api.MessageInterface {
	return &messaging_api.FlexMessage{AltText: "出勤打卡：上班／下班", Contents: &messaging_api.FlexBubble{
		Body: &messaging_api.FlexBox{Layout: "vertical", Contents: []messaging_api.FlexComponentInterface{
			&messaging_api.FlexText{Text: "出勤打卡", Weight: "bold", Size: "xl"},
			&messaging_api.FlexText{Text: "記錄上下班與休息；下班前請確認本班休息紀錄。", Wrap: true, Margin: "md"},
		}},
		Footer: &messaging_api.FlexBox{Layout: "vertical", Spacing: "md", Contents: []messaging_api.FlexComponentInterface{
			&messaging_api.FlexButton{Style: "primary", Action: &messaging_api.PostbackAction{Label: "上班", Data: "action=clock-in"}},
			&messaging_api.FlexButton{Style: "secondary", Action: &messaging_api.PostbackAction{Label: "下班", Data: "action=clock-out"}},
			&messaging_api.FlexButton{Action: &messaging_api.PostbackAction{Label: "休息開始", Data: "action=break-start"}},
			&messaging_api.FlexButton{Action: &messaging_api.PostbackAction{Label: "休息結束", Data: "action=break-end"}},
			&messaging_api.FlexButton{Action: &messaging_api.PostbackAction{Label: "確認本班休息紀錄", Data: "action=confirm-rest"}},
		}},
	}}
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
				Action: &messaging_api.PostbackAction{Label: "連線狀態", Data: "action=status"},
			},
			{
				Bounds: &messaging_api.RichMenuBounds{X: 1666, Y: 0, Width: 834, Height: 843},
				Action: &messaging_api.UriAction{Label: "開啟 MINI App", Uri: appURL},
			},
		},
	}
}
