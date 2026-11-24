package line

import (
	"github.com/963sup/line-mini-app-bot/internal/transport/line/navigation"
	"github.com/line/line-bot-sdk-go/v8/linebot/messaging_api"
)

func NavigationActions() []navigation.PrivateAction {
	return []navigation.PrivateAction{
		{Action: "clock-in", Reply: "請在上班的群組輸入「打卡」，再選擇「上班」。"},
		{Action: "transfer", Reply: "請先確認支援作業是否仍計入原班次。需要結束原班次時，請回原群組操作。此入口不會結束或新增班次。"},
		{Action: "clock-out", Reply: "請在要結束班次的原群組輸入「打卡」，再選擇「下班」。可在這裡輸入「出勤紀錄」查詢本人紀錄。"},
	}
}

// MenuActions supplies attendance actions in their published display order.
func MenuActions(group bool, journalLink string, quick bool) []messaging_api.ActionInterface {
	var actions []messaging_api.ActionInterface
	if group {
		for _, entry := range []struct{ label, command string }{
			{"打卡", "打卡"}, {"出勤", "出勤"}, {"請假登記", "請假"}, {"補卡申請", "補卡"},
		} {
			if quick {
				entry.label = entry.command
			}
			actions = append(actions, &messaging_api.MessageAction{Label: entry.label, Text: entry.command})
			if entry.command == "打卡" && journalLink != "" {
				actions = append(actions, &messaging_api.UriAction{Label: "日誌", Uri: journalLink})
			}
		}
	} else if quick {
		actions = append(actions, &messaging_api.MessageAction{Label: "出勤紀錄", Text: "出勤紀錄"})
	} else {
		actions = append(actions, &messaging_api.PostbackAction{Label: "出勤紀錄", Data: "action=attendance-history"})
	}
	return actions
}
