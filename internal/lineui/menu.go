// Package lineui renders the public connectivity menu shared by the bot and CLI.
package lineui

import "github.com/line/line-bot-sdk-go/v8/linebot/messaging_api"

func Help(group bool, appURL string) messaging_api.MessageInterface {
	title := "LINE 功能選單"
	description := "一對一連線已接通。輸入「出勤紀錄」查詢本人紀錄，或開啟網頁服務。"
	if group {
		title = "群組功能選單"
		description = "輸入「打卡」叫出上下班按鈕，也可直接輸入「上班」「下班」。歷史紀錄請私訊查詢。"
	}
	buttons := []messaging_api.FlexComponentInterface{
		&messaging_api.FlexButton{
			Action: &messaging_api.PostbackAction{Label: "連線狀態", Data: "action=status"},
		},
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

// AttendanceMenu uses silent postbacks so a click doesn't send an extra text message.
func AttendanceMenu() messaging_api.MessageInterface {
	return &messaging_api.FlexMessage{AltText: "出勤打卡：上班／下班", Contents: &messaging_api.FlexBubble{
		Body: &messaging_api.FlexBox{Layout: "vertical", Contents: []messaging_api.FlexComponentInterface{
			&messaging_api.FlexText{Text: "出勤打卡", Weight: "bold", Size: "xl"},
			&messaging_api.FlexText{Text: "點選記錄本人的上下班時間；結果會顯示在群組。", Wrap: true, Margin: "md"},
		}},
		Footer: &messaging_api.FlexBox{Layout: "horizontal", Spacing: "md", Contents: []messaging_api.FlexComponentInterface{
			&messaging_api.FlexButton{Style: "primary", Action: &messaging_api.PostbackAction{Label: "上班", Data: "action=clock-in"}},
			&messaging_api.FlexButton{Style: "secondary", Action: &messaging_api.PostbackAction{Label: "下班", Data: "action=clock-out"}},
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
