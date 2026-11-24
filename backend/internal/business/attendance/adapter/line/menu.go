package line

import "github.com/line/line-bot-sdk-go/v8/linebot/messaging_api"

// AttendanceMenu presents temporary choices for the current punch step.
func AttendanceMenu(links []string) messaging_api.MessageInterface {
	labels := []string{"上班", "下班", "休息開始", "休息結束", "確認本班休息紀錄"}
	items := make([]messaging_api.QuickReplyItem, 0, len(labels))
	for i, label := range labels {
		items = append(items, messaging_api.QuickReplyItem{Type: "action", Action: &messaging_api.UriAction{Label: label, Uri: links[i]}})
	}
	return &messaging_api.TextMessage{Text: "選擇打卡操作並開啟 MINI App；下班前請確認本班休息紀錄。", QuickReply: &messaging_api.QuickReply{Items: items}}
}
