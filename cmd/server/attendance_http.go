package main

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"mime"
	"net/http"
	"os"
	"time"

	"github.com/963sup/line-mini-app-bot/internal/attendance"
)

func registerAttendanceAPI(mux *http.ServeMux, bot *attendanceBot) {
	app := &miniApp{channelID: os.Getenv("LINE_MINI_APP_CHANNEL_ID"), client: &http.Client{Timeout: 5 * time.Second}}
	registerAttendanceQueries(mux, bot, app)
	registerAttendanceCorrections(mux, bot, app)
	registerAttendanceLaunch(mux, bot, app)
	mux.Handle("POST /api/v1/attendance/punches", http.NewCrossOriginProtection().Handler(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if bot == nil {
			writeAPIError(w, 503, "not_configured", "出勤尚未啟用。")
			return
		}
		store, ok := bot.store.(attendance.LocatedStore)
		if !ok {
			writeAPIError(w, 503, "not_configured", "定位打卡尚未啟用。")
			return
		}
		user, _, ok := app.authenticate(w, r)
		if !ok {
			return
		}
		media, _, err := mime.ParseMediaType(r.Header.Get("Content-Type"))
		if err != nil || media != "application/json" {
			writeAPIError(w, 415, "invalid_content_type", "請使用 JSON。")
			return
		}
		r.Body = http.MaxBytesReader(w, r.Body, 2048)
		decoder := json.NewDecoder(r.Body)
		decoder.DisallowUnknownFields()
		var input struct {
			RequestID string `json:"requestId"`
			locationInput
		}
		if err := decoder.Decode(&input); err != nil {
			writeAPIError(w, 400, "invalid_input", "打卡資料無效。")
			return
		}
		var extra any
		if decoder.Decode(&extra) != io.EOF {
			writeAPIError(w, 400, "invalid_input", "打卡資料無效。")
			return
		}
		now := time.Now().UTC()
		if !requestIDPattern.MatchString(input.RequestID) || !validAttendanceLocation(input.locationInput, now) {
			writeAPIError(w, 400, "invalid_position", "位置資料無效或過期，請重新取得位置。")
			return
		}
		position := attendance.Position{Latitude: *input.Latitude, Longitude: *input.Longitude, AccuracyMeters: *input.AccuracyMeters, CapturedAt: *input.CapturedAt}
		ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
		defer cancel()
		result, err := store.Submit(ctx, input.RequestID, user, position, now)
		if errors.Is(err, attendance.ErrInvalidRequest) {
			writeAPIError(w, 403, "invalid_request", "此打卡連結已過期或不屬於本人，請回原群組重新操作。")
			return
		}
		if err != nil {
			writeAPIError(w, 503, "unavailable", "暫時無法確認打卡結果，請重試或私訊查詢本人紀錄。")
			return
		}
		message := "尚無未完成的上班紀錄，未新增打卡。"
		switch result.Code {
		case "clocked_in":
			message = "上班打卡成功，" + bot.format(result.Record.ClockInAt)
		case "already_in":
			message = "已於 " + bot.format(result.Record.ClockInAt) + " 上班，無須重複打卡。"
		case "clocked_out":
			message = "下班打卡成功，" + bot.format(result.Record.ClockOutAt)
		case "break_started":
			message = "已記錄休息開始。"
		case "break_ended":
			message = "已記錄休息結束；請在下班前確認本班休息紀錄。"
		case "rest_confirmed":
			message = "已確認本班休息紀錄（未記錄休息時代表確認沒有可扣除的休息區間）。"
		case "already_resting":
			message = "休息已開始，無須重複操作。"
		case "not_resting":
			message = "沒有休息開始紀錄，未新增休息結束。"
		case "break_incomplete":
			message = "休息尚未結束，無法確認。"
		case "stale_event":
			message = "較早的打卡請求未變更出勤，請回群組重新操作。"
		}
		writeJSON(w, 200, struct {
			Code    string `json:"code"`
			Message string `json:"message"`
		}{result.Code, message})
	})))
}

func validAttendanceLocation(input locationInput, now time.Time) bool {
	return input.Latitude != nil && input.Longitude != nil && input.AccuracyMeters != nil && input.CapturedAt != nil &&
		locationNumberInRange(*input.Latitude, -90, 90) && locationNumberInRange(*input.Longitude, -180, 180) &&
		locationNumberInRange(*input.AccuracyMeters, 0, 1.7976931348623157e308) &&
		!input.CapturedAt.Before(now.Add(-5*time.Minute)) && !input.CapturedAt.After(now.Add(time.Minute))
}
