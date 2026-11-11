package line

import (
	"context"
	"fmt"
	"net/url"
	"strconv"
	"strings"
	"time"

	transport "github.com/963sup/line-mini-app-bot/internal/transport/line"

	"github.com/963sup/line-mini-app-bot/internal/attendance"
	"github.com/line/line-bot-sdk-go/v8/linebot/messaging_api"
	"github.com/line/line-bot-sdk-go/v8/linebot/webhook"
)

type Attendance struct {
	Store     attendance.Store
	Zone      *time.Location
	AppURL    string
	GroupName func(context.Context, string) string
}

// reply handles attendance before the prototype's redelivery filter.
func (b *Attendance) Reply(ctx context.Context, event webhook.EventInterface) (transport.Reply, bool, error) {
	var source webhook.SourceInterface
	var command, token, eventID, selectedGroup string
	page := 0
	month := ""
	switch e := event.(type) {
	case webhook.MessageEvent:
		if e.Mode != "active" || e.ReplyToken == "" {
			return transport.Reply{}, false, nil
		}
		m, ok := e.Message.(webhook.TextMessageContent)
		if !ok {
			return transport.Reply{}, false, nil
		}
		command = strings.TrimSpace(m.Text)
		source, token, eventID = e.Source, e.ReplyToken, e.WebhookEventId
	case webhook.PostbackEvent:
		if e.Mode != "active" || e.ReplyToken == "" || e.Postback == nil {
			return transport.Reply{}, false, nil
		}
		values, err := url.ParseQuery(e.Postback.Data)
		if err != nil || len(values["action"]) != 1 {
			return transport.Reply{}, false, nil
		}
		action := values.Get("action")
		if action == "attendance-history" {
			if len(values) > 4 {
				return transport.Reply{}, false, nil
			}
			if _, ok := values["group"]; ok {
				if len(values["group"]) != 1 || values.Get("group") == "" {
					return transport.Reply{}, false, nil
				}
				selectedGroup = values.Get("group")
			}
			for key, entries := range values {
				if len(entries) != 1 || (key != "action" && key != "group" && key != "page" && key != "month") {
					return transport.Reply{}, false, nil
				}
			}
			month = values.Get("month")
			if month != "" {
				if _, err := time.Parse("2006-01", month); err != nil || selectedGroup == "" {
					return transport.Reply{}, false, nil
				}
			}
			if _, ok := values["page"]; ok {
				var err error
				page, err = strconv.Atoi(values.Get("page"))
				if err != nil || page < 0 || len(values["page"]) != 1 {
					return transport.Reply{}, false, nil
				}
			}
		} else if len(values) != 1 {
			return transport.Reply{}, false, nil
		}
		if action != "attendance-history" {
			return transport.Reply{}, false, nil
		}
		command = "出勤紀錄"
		source, token, eventID = e.Source, e.ReplyToken, e.WebhookEventId
	default:
		return transport.Reply{}, false, nil
	}
	history := command == "出勤紀錄" || strings.HasPrefix(command, "出勤紀錄 ")
	if command != "補卡" && command != "出勤" && command != "漏下班" && command != "下班提醒" && command != "請假" && command != "打卡" && command != "上班" && command != "下班" && command != "休息開始" && command != "休息結束" && command != "確認休息" && !history {
		return transport.Reply{}, false, nil
	}
	respond := func(text string) (transport.Reply, bool, error) {
		return transport.Reply{Token: token, Message: &messaging_api.TextMessage{Text: text}}, true, nil
	}
	if b == nil {
		return respond("出勤尚未啟用。")
	}
	if user, ok := source.(webhook.UserSource); ok {
		if !history {
			if command == "補卡" {
				return respond("請回原群組輸入「補卡」取得申請入口。")
			}
			if command == "出勤" || command == "漏下班" || command == "下班提醒" {
				return respond("請回原群組輸入「出勤」或「漏下班」。")
			}
			return respond("請在群組輸入「打卡」或「請假」，取得操作入口。")
		}
		if user.UserId == "" {
			return respond("無法確認本人身分。")
		}
		if store, ok := b.Store.(attendance.GroupHistoryReader); ok {
			groups, err := store.Groups(ctx, user.UserId)
			if err != nil {
				return transport.Reply{}, true, err
			}
			if len(groups) == 0 {
				return respond("目前沒有出勤紀錄。")
			}
			if selectedGroup != "" {
				allowed := false
				for _, group := range groups {
					if group == selectedGroup {
						allowed = true
						break
					}
				}
				if !allowed {
					return respond("您沒有這個群組的出勤紀錄，請重新選擇。")
				}
			} else if command == "出勤紀錄" {
				return transport.Reply{Token: token, Message: b.historyGroups(ctx, groups, page)}, true, nil
			} else {
				number, err := strconv.Atoi(strings.TrimSpace(strings.TrimPrefix(command, "出勤紀錄 ")))
				if err != nil || number < 1 || number > len(groups) {
					return respond("群組編號無效，請輸入「出勤紀錄」重新選擇。")
				}
				selectedGroup = groups[number-1]
			}
			zone := b.Zone
			if zone == nil {
				zone = time.FixedZone("Asia/Taipei", 8*60*60)
			}
			now := time.Now().In(zone)
			start := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, zone)
			if month != "" {
				start, _ = time.ParseInLocation("2006-01", month, zone)
			}
			if month == "" {
				return transport.Reply{Token: token, Message: b.historyMonths(ctx, selectedGroup, page)}, true, nil
			}
			monthly, ok := b.Store.(interface {
				RecordsOverlappingMonth(context.Context, string, string, time.Time) ([]attendance.Record, error)
			})
			if !ok {
				return respond("月度出勤查詢尚未啟用。")
			}
			records, err := monthly.RecordsOverlappingMonth(ctx, user.UserId, selectedGroup, start)
			if err != nil {
				return transport.Reply{}, true, err
			}
			return transport.Reply{Token: token, Message: b.historyResult(ctx, selectedGroup, records, start, page)}, true, nil
		}
		records, err := b.Store.Recent(ctx, user.UserId)
		if err != nil {
			return transport.Reply{}, true, err
		}
		return respond(b.recordsText(records))
	}
	group, ok := source.(webhook.GroupSource)
	if !ok {
		return respond("請在群組打卡，或私訊查詢本人紀錄。")
	}
	if history {
		return respond("請私訊輸入「出勤紀錄」，查看本人紀錄。")
	}
	if group.UserId == "" || group.GroupId == "" {
		return respond("無法確認群組或操作人員，未執行打卡。")
	}
	if command == "補卡" {
		if b.AppURL == "" {
			return respond("補卡申請尚未啟用。")
		}
		target, err := url.Parse(b.AppURL)
		if err != nil {
			return transport.Reply{}, true, err
		}
		query := target.Query()
		query.Set("correction", group.GroupId)
		target.RawQuery = query.Encode()
		return transport.Reply{Token: token, Message: &messaging_api.TemplateMessage{AltText: "補卡／時間更正申請：一律須人工審核。", Template: &messaging_api.ButtonsTemplate{Text: "漏上班、漏下班或整班皆漏，請申報已結束班次的實際時間。核准前有效紀錄不變。", Actions: []messaging_api.ActionInterface{&messaging_api.UriAction{Label: "補卡／時間更正", Uri: target.String()}}}}}, true, nil
	}

	if command == "出勤" || command == "漏下班" || command == "下班提醒" {
		return b.replyGroupStatus(ctx, group.GroupId, group.UserId, command, token)
	}
	if command == "請假" {
		store, ok := b.Store.(attendance.LeaveStore)
		if !ok || b.AppURL == "" {
			return respond("請假尚未啟用。")
		}
		launch, err := store.CreateLeaveLaunch(ctx, group.GroupId)
		if err != nil {
			return transport.Reply{}, true, err
		}
		link := b.AppURL + "?leave=" + url.QueryEscape(launch)
		return transport.Reply{Token: token, Message: &messaging_api.TemplateMessage{
			AltText:  "請假登記：開啟表單填寫，送出不代表核准。",
			Template: &messaging_api.ButtonsTemplate{Text: "請假登記（30分鐘內開啟並送出）。理由不會發送到群組；送出不代表核准。", Actions: []messaging_api.ActionInterface{&messaging_api.UriAction{Label: "填寫請假", Uri: link}}},
		}}, true, nil
	}
	if command != "打卡" {
		return respond("請重新輸入「打卡」取得最新操作選單。")
	}
	store, ok := b.Store.(attendance.LaunchStore)
	if !ok || b.AppURL == "" || eventID == "" {
		return respond("定位打卡頁尚未設定，未執行打卡。")
	}
	actions := []string{"clock-in", "clock-out", "break-start", "break-end", "confirm-rest"}
	links := make([]string, 0, len(actions))
	for _, action := range actions {
		launch, err := store.CreateLaunch(ctx, group.GroupId, action)
		if err != nil {
			return transport.Reply{}, true, err
		}
		links = append(links, b.AppURL+"?launch="+url.QueryEscape(launch))
	}
	return transport.Reply{Token: token, Message: AttendanceMenu(links)}, true, nil
}

func (b *Attendance) recordsText(records []attendance.Record) string {
	text := "本人最近出勤紀錄（最多五筆）："
	for _, record := range records {
		out := "尚未下班"
		if !record.ClockOutAt.IsZero() {
			out = b.format(record.ClockOutAt)
		}
		text += fmt.Sprintf("\n群組 %s\n上班 %s\n下班 %s", record.GroupID, b.format(record.ClockInAt), out)
	}
	if len(records) == 0 {
		return "目前沒有出勤紀錄。"
	}
	return text
}

func (b *Attendance) format(t time.Time) string { return t.In(b.Zone).Format("2006-01-02 15:04:05") }
