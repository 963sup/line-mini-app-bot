package line

import (
	"context"
	"errors"
	"fmt"
	"net/url"
	"time"

	attendancedto "github.com/963sup/line-mini-app-bot/internal/business/attendance/application/dto"
	transport "github.com/963sup/line-mini-app-bot/internal/transport/line"
	"github.com/line/line-bot-sdk-go/v8/linebot/messaging_api"
)

func (b *Attendance) groupStatus(ctx context.Context, group, user, command string) (messaging_api.MessageInterface, error) {
	text := func(value string) messaging_api.MessageInterface { return &messaging_api.TextMessage{Text: value} }
	if b.Store == nil {
		return text("出勤查詢與漏下班檢查尚未啟用。"), nil
	}
	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()
	if command == "出勤" {
		now := time.Now().UTC()
		zone := b.Zone
		if zone == nil {
			zone = time.FixedZone("Asia/Taipei", 8*60*60)
		}
		day := now.In(zone)
		start := time.Date(day.Year(), day.Month(), day.Day(), 0, 0, 0, 0, zone)
		board, err := b.Store.GroupBoard(ctx, group, start, now)
		if errors.Is(err, attendancedto.ErrBoardLimit) {
			return text("資料超出目前看板讀取上限，無法取得完整人數；本次不顯示不完整統計。"), nil
		}
		if err != nil {
			return nil, err
		}
		return &messaging_api.FlexMessage{AltText: "出勤", Contents: &messaging_api.FlexBubble{Body: &messaging_api.FlexBox{Layout: "vertical", Contents: []messaging_api.FlexComponentInterface{
			&messaging_api.FlexText{Text: "出勤", Weight: "bold", Size: "xl"},
			&messaging_api.FlexText{Text: day.Format("2006-01-02") + "｜臺灣時間", Color: "#666666", Margin: "sm"},
			&messaging_api.FlexText{Text: fmt.Sprintf("上班中　%d 人\n休息中　%d 人\n今日已下班　%d 人", board.Working, board.Resting, board.Finished), Wrap: true, Margin: "md"},
			&messaging_api.FlexText{Text: "更新：" + board.AsOf.In(zone).Format("2006-01-02 15:04:05") + "\n依打卡紀錄計算；跨日未結束班次仍算上班／休息中。未打卡不代表缺勤。輸入「漏下班」檢查本人未結束班次。", Wrap: true, Size: "sm", Color: "#666666", Margin: "md"},
		}}}}, nil
	}
	active, err := b.Store.CurrentShift(ctx, group, user)
	if err != nil {
		return nil, err
	}
	if active == nil {
		return text("本次查詢者在本群沒有未結束班次。"), nil
	}
	if b.AppURL == "" {
		return text("本次查詢者仍有未結束班次。現在下班請回群組輸入「打卡」；過去漏打請私訊「出勤紀錄」提出更正。"), nil
	}
	launch, err := b.Store.CreateLaunch(ctx, group, "clock-out")
	if err != nil {
		return nil, err
	}
	return &messaging_api.TemplateMessage{
		AltText: "本人漏下班檢查：仍有未結束班次；現在下班或私訊查詢並更正。",
		Template: &messaging_api.ButtonsTemplate{
			Text:    "查詢者仍有未結束班次。現在下班會記錄目前時間；過去漏打請私訊「出勤紀錄」更正。",
			Actions: []messaging_api.ActionInterface{&messaging_api.UriAction{Label: "現在下班（需定位）", Uri: b.AppURL + "?launch=" + url.QueryEscape(launch)}},
		},
	}, nil
}

func (b *Attendance) replyGroupStatus(ctx context.Context, group, user, command, token string) (transport.Reply, bool, error) {
	message, err := b.groupStatus(ctx, group, user, command)
	return transport.Reply{Token: token, Message: message}, true, err
}
