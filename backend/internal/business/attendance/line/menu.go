package line

import "github.com/line/line-bot-sdk-go/v8/linebot/messaging_api"

// AttendanceMenu uses silent postbacks so a click doesn't send an extra text message.
func AttendanceMenu(links []string) messaging_api.MessageInterface {
	labels := []string{"上班", "下班", "休息開始", "休息結束", "確認本班休息紀錄"}
	buttons := make([]messaging_api.FlexComponentInterface, 0, len(labels))
	for i, label := range labels {
		buttons = append(buttons, &messaging_api.FlexButton{Action: &messaging_api.UriAction{Label: label, Uri: links[i]}})
	}
	return &messaging_api.FlexMessage{AltText: "出勤打卡：上班／下班", Contents: &messaging_api.FlexBubble{
		Body: &messaging_api.FlexBox{Layout: "vertical", Contents: []messaging_api.FlexComponentInterface{
			&messaging_api.FlexText{Text: "出勤打卡", Weight: "bold", Size: "xl"},
			&messaging_api.FlexText{Text: "選擇操作並開啟 MINI App；下班前請確認本班休息紀錄。", Wrap: true, Margin: "md"},
		}},
		Footer: &messaging_api.FlexBox{Layout: "vertical", Spacing: "md", Contents: buttons},
	}}
}
