package line

import (
	"context"
	"fmt"
	"net/url"
	"strings"
	"time"

	"github.com/963sup/line-mini-app-bot/internal/attendance"
	"github.com/963sup/line-mini-app-bot/internal/lineui"
	"github.com/line/line-bot-sdk-go/v8/linebot/messaging_api"
)

func (b *Attendance) HistoryGroupName(ctx context.Context, group string) string {
	name := ""
	if b.GroupName != nil {
		name = b.GroupName(ctx, group)
	}
	return lineui.GroupName(group, name)
}

func historyButton(label, data string) messaging_api.FlexComponentInterface {
	return &messaging_api.FlexButton{Action: &messaging_api.PostbackAction{Label: label, Data: data}}
}

func historyCard(title, text string, buttons []messaging_api.FlexComponentInterface) messaging_api.MessageInterface {
	return &messaging_api.FlexMessage{AltText: title, Contents: &messaging_api.FlexBubble{
		Body: &messaging_api.FlexBox{Layout: "vertical", Contents: []messaging_api.FlexComponentInterface{
			&messaging_api.FlexText{Text: title, Weight: "bold", Wrap: true},
			&messaging_api.FlexText{Text: text, Wrap: true, Margin: "md"},
		}}, Footer: &messaging_api.FlexBox{Layout: "vertical", Contents: buttons},
	}}
}

func (b *Attendance) historyGroups(ctx context.Context, groups []string, page int) messaging_api.MessageInterface {
	const size = 5
	if page > (len(groups)-1)/size {
		page = 0
	}
	end := (page + 1) * size
	if end > len(groups) {
		end = len(groups)
	}
	buttons := []messaging_api.FlexComponentInterface{}
	var text strings.Builder
	text.WriteString("點選按鈕，接著選擇月份，再查看本人紀錄。\n")
	for i := page * size; i < end; i++ {
		fmt.Fprintf(&text, "\n%d. %s", i+1, b.HistoryGroupName(ctx, groups[i]))
		data := url.Values{"action": {"attendance-history"}, "group": {groups[i]}}.Encode()
		buttons = append(buttons, historyButton(fmt.Sprintf("查看群組 %d", i+1), data))
	}
	text.WriteString("\n\n也可輸入「出勤紀錄 編號」。")
	if page > 0 {
		buttons = append(buttons, historyButton("上一頁", fmt.Sprintf("action=attendance-history&page=%d", page-1)))
	}
	if end < len(groups) {
		buttons = append(buttons, historyButton("下一頁", fmt.Sprintf("action=attendance-history&page=%d", page+1)))
	}
	return historyCard("選擇出勤群組", text.String(), buttons)
}

func (b *Attendance) historyResult(ctx context.Context, group string, records []attendance.Record, start time.Time, page int) messaging_api.MessageInterface {
	zone := start.Location()
	const size = 7
	pages := (len(records) + size - 1) / size
	if pages == 0 {
		pages = 1
	}
	if page >= pages {
		page = 0
	}
	end := (page + 1) * size
	if end > len(records) {
		end = len(records)
	}
	calculated := attendance.CalculateMonthlyAttendance(records, start).Calculation
	var text strings.Builder
	fmt.Fprintf(&text, "%s｜臺灣時間\n本月已確認工作 %d小時%d分\n待確認 %d 筆・異常 %d 項\n正常／加班分類待認定\n第 %d/%d 頁\n", start.Format("2006年01月"), calculated.Minutes/60, calculated.Minutes%60, calculated.Pending, len(calculated.Issues), page+1, pages)
	for _, r := range records[page*size : end] {
		out := "尚未下班"
		if !r.ClockOutAt.IsZero() {
			out = r.ClockOutAt.In(zone).Format("01/02 15:04")
		}
		fmt.Fprintf(&text, "\n%s 上班 %s\n下班 %s\n", r.ClockInAt.In(zone).Format("01/02"), r.ClockInAt.In(zone).Format("15:04"), out)
		if !r.BreakStartedAt.IsZero() {
			text.WriteString("休息未結束・待確認\n")
		} else if !r.RestConfirmed {
			text.WriteString("休息未確認・未納入工時\n")
		}
	}
	if len(records) == 0 {
		text.WriteString("\n本月沒有出勤紀錄。")
	}
	text.WriteString("\n跨月班次只計本月區間；休息未確認與漏卡不納入已確認工時。")
	action := func(month time.Time, p int) string {
		return url.Values{"action": {"attendance-history"}, "group": {group}, "month": {month.Format("2006-01")}, "page": {fmt.Sprint(p)}}.Encode()
	}
	buttons := []messaging_api.FlexComponentInterface{}
	if page > 0 {
		buttons = append(buttons, historyButton("上一頁", action(start, page-1)))
	}
	if page+1 < pages {
		buttons = append(buttons, historyButton("下一頁", action(start, page+1)))
	}
	buttons = append(buttons, historyButton("上個月", action(start.AddDate(0, -1, 0), 0)), historyButton("下個月", action(start.AddDate(0, 1, 0), 0)), historyButton("選擇月份", url.Values{"action": {"attendance-history"}, "group": {group}}.Encode()), historyButton("切換群組", "action=attendance-history&page=0"))
	return historyCard(b.HistoryGroupName(ctx, group)+"｜本人月度出勤", text.String(), buttons)
}

func (b *Attendance) historyMonths(ctx context.Context, group string, page int) messaging_api.MessageInterface {
	if page > 199 {
		page = 0
	}
	zone := b.Zone
	if zone == nil {
		zone = time.FixedZone("Asia/Taipei", 8*60*60)
	}
	now := time.Now().In(zone)
	current := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, zone)
	buttons := []messaging_api.FlexComponentInterface{}
	for i := 0; i < 6; i++ {
		month := current.AddDate(0, -page*6-i, 0)
		data := url.Values{"action": {"attendance-history"}, "group": {group}, "month": {month.Format("2006-01")}}.Encode()
		buttons = append(buttons, historyButton(month.Format("2006年01月"), data))
	}
	pageData := func(p int) string {
		return url.Values{"action": {"attendance-history"}, "group": {group}, "page": {fmt.Sprint(p)}}.Encode()
	}
	if page > 0 {
		buttons = append(buttons, historyButton("較新月份", pageData(page-1)))
	}
	if page < 199 {
		buttons = append(buttons, historyButton("更早月份", pageData(page+1)))
	}
	buttons = append(buttons, historyButton("切換群組", "action=attendance-history&page=0"))
	return historyCard(b.HistoryGroupName(ctx, group)+"｜選擇月份", "請選擇要查詢的月份，僅顯示本人的出勤紀錄。", buttons)
}
