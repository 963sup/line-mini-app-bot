package line

import (
	"context"
	"fmt"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/963sup/line-mini-app-bot/internal/attendance"
	"github.com/963sup/line-mini-app-bot/internal/lineui"
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
func (b *Attendance) reply(ctx context.Context, event webhook.EventInterface) (lineReply, bool, error) {
	var source webhook.SourceInterface
	var command, token, eventID, selectedGroup string
	page := 0
	month := ""
	switch e := event.(type) {
	case webhook.MessageEvent:
		if e.Mode != "active" || e.ReplyToken == "" {
			return lineReply{}, false, nil
		}
		m, ok := e.Message.(webhook.TextMessageContent)
		if !ok {
			return lineReply{}, false, nil
		}
		command = strings.TrimSpace(m.Text)
		source, token, eventID = e.Source, e.ReplyToken, e.WebhookEventId
	case webhook.PostbackEvent:
		if e.Mode != "active" || e.ReplyToken == "" || e.Postback == nil {
			return lineReply{}, false, nil
		}
		values, err := url.ParseQuery(e.Postback.Data)
		if err != nil || len(values["action"]) != 1 {
			return lineReply{}, false, nil
		}
		action := values.Get("action")
		if action == "attendance-history" {
			if len(values) > 4 {
				return lineReply{}, false, nil
			}
			if _, ok := values["group"]; ok {
				if len(values["group"]) != 1 || values.Get("group") == "" {
					return lineReply{}, false, nil
				}
				selectedGroup = values.Get("group")
			}
			for key, entries := range values {
				if len(entries) != 1 || (key != "action" && key != "group" && key != "page" && key != "month") {
					return lineReply{}, false, nil
				}
			}
			month = values.Get("month")
			if month != "" {
				if _, err := time.Parse("2006-01", month); err != nil || selectedGroup == "" {
					return lineReply{}, false, nil
				}
			}
			if _, ok := values["page"]; ok {
				var err error
				page, err = strconv.Atoi(values.Get("page"))
				if err != nil || page < 0 || len(values["page"]) != 1 {
					return lineReply{}, false, nil
				}
			}
		} else if len(values) != 1 {
			return lineReply{}, false, nil
		}
		if action != "attendance-history" {
			return lineReply{}, false, nil
		}
		command = "出勤紀錄"
		source, token, eventID = e.Source, e.ReplyToken, e.WebhookEventId
	default:
		return lineReply{}, false, nil
	}
	history := command == "出勤紀錄" || strings.HasPrefix(command, "出勤紀錄 ")
	if command != "打卡" && command != "上班" && command != "下班" && command != "休息開始" && command != "休息結束" && command != "確認休息" && !history {
		return lineReply{}, false, nil
	}
	respond := func(text string) (lineReply, bool, error) {
		return lineReply{token: token, message: &messaging_api.TextMessage{Text: text}}, true, nil
	}
	if b == nil {
		return respond("出勤尚未啟用。")
	}
	if user, ok := source.(webhook.UserSource); ok {
		if !history {
			return respond("請在群組輸入「打卡」，點選上班或下班。")
		}
		if user.UserId == "" {
			return respond("無法確認本人身分。")
		}
		if store, ok := b.Store.(attendance.LocatedStore); ok {
			groups, err := store.Groups(ctx, user.UserId)
			if err != nil {
				return lineReply{}, true, err
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
				return lineReply{token: token, message: b.historyGroups(ctx, groups, page)}, true, nil
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
				return lineReply{token: token, message: b.historyMonths(ctx, selectedGroup, page)}, true, nil
			}
			monthly, ok := b.Store.(interface {
				MonthGroup(context.Context, string, string, time.Time) ([]attendance.Record, error)
			})
			if !ok {
				return respond("月度出勤查詢尚未啟用。")
			}
			records, err := monthly.MonthGroup(ctx, user.UserId, selectedGroup, start)
			if err != nil {
				return lineReply{}, true, err
			}
			return lineReply{token: token, message: b.historyResult(ctx, selectedGroup, records, start, page)}, true, nil
		}
		records, err := b.Store.Recent(ctx, user.UserId)
		if err != nil {
			return lineReply{}, true, err
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
			return lineReply{}, true, err
		}
		links = append(links, b.AppURL+"?launch="+url.QueryEscape(launch))
	}
	return lineReply{token: token, message: lineui.AttendanceMenu(links)}, true, nil
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
