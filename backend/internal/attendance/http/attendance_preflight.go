package attendancehttp

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"mime"
	"net/http"
	"time"

	"github.com/963sup/line-mini-app-bot/internal/attendance"
	"github.com/963sup/line-mini-app-bot/internal/httpapi/generated"
	transport "github.com/963sup/line-mini-app-bot/internal/transport/http"
)

func registerPunchPreflight(mux *http.ServeMux, bot *Attendance, app Authenticator) {
	mux.Handle("POST /api/v1/attendance/punches/preflight", http.NewCrossOriginProtection().Handler(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if bot == nil {
			transport.WriteAPIError(w, 503, "not_configured", "出勤尚未啟用。")
			return
		}
		store, ok := bot.store.(attendance.PunchResolver)
		if !ok {
			transport.WriteAPIError(w, 503, "not_configured", "打卡狀態確認尚未啟用。")
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
		r.Body = http.MaxBytesReader(w, r.Body, 512)
		decoder := json.NewDecoder(r.Body)
		decoder.DisallowUnknownFields()
		var input generated.PunchPreflightRequest
		if decoder.Decode(&input) != nil || decoder.Decode(new(any)) != io.EOF || !requestIDPattern.MatchString(input.RequestId) {
			transport.WriteAPIError(w, 400, "invalid_input", "打卡資料無效。")
			return
		}
		ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
		defer cancel()
		result, err := store.ResolvePunchAttempt(ctx, input.RequestId, user, time.Now().UTC())
		if errors.Is(err, attendance.ErrInvalidRequest) {
			transport.WriteAPIError(w, 403, "invalid_request", "此打卡連結已過期或不屬於本人，請回原群組重新操作。")
			return
		}
		if err != nil {
			transport.WriteAPIError(w, 503, "unavailable", "暫時無法確認打卡狀態，請重試或私訊查詢本人紀錄。")
			return
		}
		transport.WriteJSON(w, 200, bot.punchResponse(result))
	})))
}
