package main

import (
	"context"
	"fmt"
	"net/url"
	"os"
	"regexp"
	"strings"
	"time"
	_ "time/tzdata"

	"cloud.google.com/go/firestore"
	storage "github.com/963sup/line-mini-app-bot/internal/adapters/firestore"
	"github.com/963sup/line-mini-app-bot/internal/attendance"
	"github.com/963sup/line-mini-app-bot/internal/lineui"
	"github.com/line/line-bot-sdk-go/v8/linebot/messaging_api"
	"github.com/line/line-bot-sdk-go/v8/linebot/webhook"
)

type attendanceBot struct {
	store attendance.Store
	zone  *time.Location
}

func newAttendanceBot(ctx context.Context, channel string) (*attendanceBot, func() error, error) {
	project := os.Getenv("GOOGLE_CLOUD_PROJECT")
	if project == "" || channel == "" {
		return nil, func() error { return nil }, nil
	}
	if !regexp.MustCompile(`^[0-9]+$`).MatchString(channel) {
		return nil, nil, fmt.Errorf("attendance requires LINE_MESSAGING_API_CHANNEL_ID")
	}
	database := os.Getenv("FIRESTORE_DATABASE_ID")
	if database == "" {
		database = "(default)"
	}
	client, err := firestore.NewClientWithDatabase(ctx, project, database)
	if err != nil {
		return nil, nil, fmt.Errorf("cannot initialize attendance database")
	}
	zone, err := time.LoadLocation("Asia/Taipei")
	if err != nil {
		client.Close()
		return nil, nil, err
	}
	return &attendanceBot{store: &storage.Attendance{Client: client, ChannelID: channel}, zone: zone}, client.Close, nil
}

// reply handles attendance before the prototype's redelivery filter.
func (b *attendanceBot) reply(ctx context.Context, event webhook.EventInterface) (lineReply, bool, error) {
	var source webhook.SourceInterface
	var command, token, eventID string
	var timestamp int64
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
		source, token, eventID, timestamp = e.Source, e.ReplyToken, e.WebhookEventId, e.Timestamp
	case webhook.PostbackEvent:
		if e.Mode != "active" || e.ReplyToken == "" || e.Postback == nil {
			return lineReply{}, false, nil
		}
		values, err := url.ParseQuery(e.Postback.Data)
		if err != nil || len(values) != 1 || len(values["action"]) != 1 {
			return lineReply{}, false, nil
		}
		switch values.Get("action") {
		case "clock-in":
			command = "上班"
		case "clock-out":
			command = "下班"
		default:
			return lineReply{}, false, nil
		}
		source, token, eventID, timestamp = e.Source, e.ReplyToken, e.WebhookEventId, e.Timestamp
	default:
		return lineReply{}, false, nil
	}
	if command != "打卡" && command != "上班" && command != "下班" && command != "出勤紀錄" {
		return lineReply{}, false, nil
	}
	respond := func(text string) (lineReply, bool, error) {
		return lineReply{token: token, message: &messaging_api.TextMessage{Text: text}}, true, nil
	}
	if b == nil {
		return respond("出勤尚未啟用。")
	}
	if user, ok := source.(webhook.UserSource); ok {
		if command != "出勤紀錄" {
			return respond("請在群組輸入「打卡」，點選上班或下班。")
		}
		if user.UserId == "" {
			return respond("無法確認本人身分。")
		}
		records, err := b.store.Recent(ctx, user.UserId)
		if err != nil {
			return lineReply{}, true, err
		}
		text := "本人最近出勤紀錄（最多五筆）："
		for _, record := range records {
			out := "尚未下班"
			if !record.ClockOutAt.IsZero() {
				out = b.format(record.ClockOutAt)
			}
			text += fmt.Sprintf("\n群組 %s\n上班 %s\n下班 %s", record.GroupID, b.format(record.ClockInAt), out)
		}
		if len(records) == 0 {
			text = "目前沒有出勤紀錄。"
		}
		return respond(text)
	}
	group, ok := source.(webhook.GroupSource)
	if !ok {
		return respond("請在群組打卡，或私訊查詢本人紀錄。")
	}
	if command == "出勤紀錄" {
		return respond("請私訊輸入「出勤紀錄」，查看本人紀錄。")
	}
	if group.UserId == "" || group.GroupId == "" {
		return respond("無法確認群組或操作人員，未執行打卡。")
	}
	if command == "打卡" {
		return lineReply{token: token, message: lineui.AttendanceMenu()}, true, nil
	}
	if eventID == "" {
		return respond("事件缺少識別碼，未執行打卡。")
	}
	action := "clock-in"
	if command == "下班" {
		action = "clock-out"
	}
	result, err := b.store.Punch(ctx, group.GroupId, group.UserId, eventID, action, time.Now().UTC(), timestamp)
	if err != nil {
		return lineReply{}, true, err
	}
	switch result.Code {
	case "stale_event":
		return respond("收到較早的延遲事件，未變更出勤；請確認紀錄後重新操作。")
	case "clocked_in":
		return respond("上班打卡成功，" + b.format(result.Record.ClockInAt))
	case "already_in":
		return respond("已於 " + b.format(result.Record.ClockInAt) + " 上班，無須重複打卡。")
	case "clocked_out":
		return respond("下班打卡成功，" + b.format(result.Record.ClockOutAt))
	default:
		return respond("尚無未完成的上班紀錄，未新增打卡。")
	}
}

func (b *attendanceBot) format(t time.Time) string { return t.In(b.zone).Format("2006-01-02 15:04:05") }
