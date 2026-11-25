package attendancehttp

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"mime"
	"net/http"
	"regexp"
	"time"

	attendanceport "github.com/963sup/line-mini-app-bot/internal/business/attendance/application/port"
	transport "github.com/963sup/line-mini-app-bot/internal/transport/http"
	"github.com/963sup/line-mini-app-bot/internal/transport/http/api/generated"
)

var launchTokenPattern = regexp.MustCompile(`^[0-9a-f]{64}$`)

func (s *attendanceServer) ExchangeAttendanceLaunch(w http.ResponseWriter, r *http.Request, _ generated.ExchangeAttendanceLaunchParams) {
	bot, app := s.bot, s.app

	if bot == nil {
		transport.WriteAPIError(w, 503, "not_configured", "出勤尚未啟用。")
		return
	}
	store := bot.store
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
	var input generated.LaunchRequest
	dec := json.NewDecoder(r.Body)
	dec.DisallowUnknownFields()
	if dec.Decode(&input) != nil || dec.Decode(new(any)) != io.EOF || !launchTokenPattern.MatchString(input.Launch) {
		transport.WriteAPIError(w, 400, "invalid_launch", "打卡連結無效。")
		return
	}
	ctx, cancel := context.WithTimeout(r.Context(), 8*time.Second)
	defer cancel()
	group, err := store.LaunchGroup(ctx, input.Launch)
	if errors.Is(err, attendanceport.ErrExpiredLaunch) {
		transport.WriteAPIError(w, 403, "expired_launch", "打卡連結已過期。")
		return
	}
	if err != nil {
		transport.WriteAPIError(w, 503, "unavailable", "暫時無法確認打卡連結。")
		return
	}
	if !bot.requireGroupMember(w, ctx, group, user) {
		return
	}
	id, err := store.ExchangeLaunch(ctx, input.Launch, user)
	if errors.Is(err, attendanceport.ErrExpiredLaunch) {
		transport.WriteAPIError(w, 403, "expired_launch", "打卡連結已過期。")
		return
	}
	if err != nil {
		transport.WriteAPIError(w, 503, "unavailable", "暫時無法建立打卡操作。")
		return
	}
	transport.WriteJSON(w, 200, generated.LaunchResponse{RequestId: id})
}

// Membership proof is shared by launch, leave and missing-shift submissions.
func (bot *Attendance) requireGroupMember(w http.ResponseWriter, ctx context.Context, group, user string) bool {
	if bot.membership == nil {
		transport.WriteAPIError(w, 503, "unavailable", "暫時無法確認群組身分，未送出申請。")
		return false
	}
	member, err := bot.membership.IsMember(ctx, group, user)
	if err == nil && member {
		return true
	}
	if err == nil {
		transport.WriteAPIError(w, 403, "not_member", "無法確認你是此群組成員。")
	} else {
		transport.WriteAPIError(w, 503, "unavailable", "暫時無法確認群組身分，未送出申請。")
	}
	return false
}
