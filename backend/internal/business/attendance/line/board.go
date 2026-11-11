package line

import (
	"context"
	"errors"
	"fmt"
	"net/url"
	"time"

	boardmodel "github.com/963sup/line-mini-app-bot/internal/business/attendance/board"
	"github.com/963sup/line-mini-app-bot/internal/business/attendance/punch"
	transport "github.com/963sup/line-mini-app-bot/internal/transport/line"
	"github.com/line/line-bot-sdk-go/v8/linebot/messaging_api"
)

func (b *Attendance) groupStatus(ctx context.Context, group, user, command string) (messaging_api.MessageInterface, error) {
	text := func(value string) messaging_api.MessageInterface { return &messaging_api.TextMessage{Text: value} }
	store, ok := b.Store.(boardmodel.BoardStore)
	if !ok {
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
		board, err := store.GroupBoard(ctx, group, start, now)
		if errors.Is(err, boardmodel.ErrBoardLimit) {
			return text("資料超出目前看板讀取上限，無法取得完整人數；本次不顯示不完整統計。"), nil
		}
		if err != nil {
			return nil, err
		}
		return text(fmt.Sprintf("本群出勤（%s）\n上班中：%d 人\n休息中：%d 人\n今日已下班：%d 人\n更新：%s\n依打卡紀錄計算；跨日未結束班次仍算上班／休息中。未打卡不代表缺勤。\n輸入「漏下班」檢查本人未結束班次。", day.Format("2006-01-02"), board.Working, board.Resting, board.Finished, board.AsOf.In(zone).Format("2006-01-02 15:04:05"))), nil
	}
	active, err := store.CurrentShift(ctx, group, user)
	if err != nil {
		return nil, err
	}
	if active == nil {
		return text("本次查詢者在本群沒有未結束班次。"), nil
	}
	launchStore, ok := b.Store.(punch.LaunchStore)
	if !ok || b.AppURL == "" {
		return text("本次查詢者仍有未結束班次。現在下班請回群組輸入「打卡」；過去漏打請私訊「出勤紀錄」提出更正。"), nil
	}
	launch, err := launchStore.CreateLaunch(ctx, group, "clock-out")
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
