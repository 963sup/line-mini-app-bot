package attendancehttp

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"mime"
	"net/http"
	"time"

	transport "github.com/963sup/line-mini-app-bot/internal/transport/http"

	"github.com/963sup/line-mini-app-bot/internal/attendance"
	"github.com/963sup/line-mini-app-bot/internal/httpapi/generated"
)

func RegisterAttendanceAPI(public *http.ServeMux, bot *Attendance, app Authenticator) {
	registerAttendanceAPIWithIdentity(public, bot, app)
}

func registerAttendanceAPIWithIdentity(public *http.ServeMux, bot *Attendance, app Authenticator) {
	mux := http.NewServeMux()
	registerAttendanceQueries(mux, bot, app)
	registerAttendanceCorrections(mux, bot, app)
	registerAttendanceLaunch(mux, bot, app)
	registerPunchPreflight(mux, bot, app)
	registerLeave(mux, bot, app)
	mux.Handle("POST /api/v1/attendance/punches", http.NewCrossOriginProtection().Handler(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if bot == nil {
			transport.WriteAPIError(w, 503, "not_configured", "出勤尚未啟用。")
			return
		}
		store, ok := bot.store.(attendance.PunchSubmitter)
		if !ok {
			transport.WriteAPIError(w, 503, "not_configured", "定位打卡尚未啟用。")
			return
		}
		user, _, ok := app.Authenticate(w, r)
		if !ok {
			return
		}
		media, _, err := mime.ParseMediaType(r.Header.Get("Content-Type"))
		if err != nil || media != "application/json" {
			transport.WriteAPIError(w, 415, "invalid_content_type", "請使用 JSON。")
			return
		}
		r.Body = http.MaxBytesReader(w, r.Body, 2048)
		decoder := json.NewDecoder(r.Body)
		decoder.DisallowUnknownFields()
		var input generated.PunchRequest
		if err := decoder.Decode(&input); err != nil {
			transport.WriteAPIError(w, 400, "invalid_input", "打卡資料無效。")
			return
		}
		var extra any
		if decoder.Decode(&extra) != io.EOF {
			transport.WriteAPIError(w, 400, "invalid_input", "打卡資料無效。")
			return
		}
		now := time.Now().UTC()
		if !requestIDPattern.MatchString(input.RequestId) || !validAttendanceLocation(locationInput{Latitude: &input.Latitude, Longitude: &input.Longitude, AccuracyMeters: &input.AccuracyMeters, CapturedAt: &input.CapturedAt}, now) {
			transport.WriteAPIError(w, 400, "invalid_position", "位置資料無效或過期，請重新取得位置。")
			return
		}
		position := attendance.Position{Latitude: input.Latitude, Longitude: input.Longitude, AccuracyMeters: input.AccuracyMeters, CapturedAt: input.CapturedAt}
		ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
		defer cancel()
		result, err := store.SubmitPunch(ctx, input.RequestId, user, position, now)
		if errors.Is(err, attendance.ErrInvalidRequest) {
			transport.WriteAPIError(w, 403, "invalid_request", "此打卡連結已過期或不屬於本人，請回原群組重新操作。")
			return
		}
		if err != nil {
			transport.WriteAPIError(w, 503, "unavailable", "暫時無法確認打卡結果，請重試或私訊查詢本人紀錄。")
			return
		}
		transport.WriteJSON(w, 200, bot.punchResponse(result))
	})))
	registerContractAttendance(public, mux, bot, app)
}

func validAttendanceLocation(input locationInput, now time.Time) bool {
	return input.Latitude != nil && input.Longitude != nil && input.AccuracyMeters != nil && input.CapturedAt != nil &&
		locationNumberInRange(*input.Latitude, -90, 90) && locationNumberInRange(*input.Longitude, -180, 180) &&
		locationNumberInRange(*input.AccuracyMeters, 0, 1.7976931348623157e308) &&
		!input.CapturedAt.Before(now.Add(-5*time.Minute)) && !input.CapturedAt.After(now.Add(time.Minute))
}

func (bot *Attendance) punchResponse(result attendance.Result) generated.PunchResponse {
	message := "尚無未完成的上班紀錄，未新增打卡。"
	switch result.Code {
	case "needs_position":
		message = "需要取得目前位置才能完成此操作。"
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
	return generated.PunchResponse{Code: result.Code, Message: message}
}
