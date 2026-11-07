// Package lineui renders the public connectivity menu shared by the bot and CLI.
package lineui

import "github.com/line/line-bot-sdk-go/v8/linebot/messaging_api"

func Help(group bool, appURL string) messaging_api.MessageInterface {
	title := "LINE 功能選單"
	description := "一對一連線已接通。可查看連線狀態，或開啟網頁服務；出勤尚未啟用。"
	if group {
		title = "群組功能選單"
		description = "群組只提供公開功能說明與連線狀態。個人資料請至一對一或網頁查看。"
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
