package main

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"mime"
	"net/http"
	"net/url"
	"os"
	"regexp"
	"time"

	storage "github.com/963sup/line-mini-app-bot/internal/adapters/firestore"
	"github.com/963sup/line-mini-app-bot/internal/attendance"
)

var launchTokenPattern = regexp.MustCompile(`^[0-9a-f]{64}$`)

func registerAttendanceLaunch(mux *http.ServeMux, bot *attendanceBot, app *miniApp) {
	mux.Handle("POST /api/v1/attendance/launch", http.NewCrossOriginProtection().Handler(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if bot == nil {
			writeAPIError(w, 503, "not_configured", "出勤尚未啟用。")
			return
		}
		store, ok := bot.store.(attendance.LaunchStore)
		if !ok {
			writeAPIError(w, 503, "not_configured", "打卡連結尚未啟用。")
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
		r.Body = http.MaxBytesReader(w, r.Body, 512)
		var input struct {
			Launch string `json:"launch"`
		}
		dec := json.NewDecoder(r.Body)
		dec.DisallowUnknownFields()
		if dec.Decode(&input) != nil || dec.Decode(new(any)) != io.EOF || !launchTokenPattern.MatchString(input.Launch) {
			writeAPIError(w, 400, "invalid_launch", "打卡連結無效。")
			return
		}
		ctx, cancel := context.WithTimeout(r.Context(), 8*time.Second)
		defer cancel()
		group, err := store.LaunchGroup(ctx, input.Launch)
		if errors.Is(err, storage.ErrExpiredLaunch) {
			writeAPIError(w, 403, "expired_launch", "打卡連結已過期。")
			return
		}
		if err != nil {
			writeAPIError(w, 503, "unavailable", "暫時無法確認打卡連結。")
			return
		}
		client := &http.Client{Timeout: 5 * time.Second, CheckRedirect: func(*http.Request, []*http.Request) error { return http.ErrUseLastResponse }}
		memberStatus := checkGroupMember(ctx, client, "https://api.line.me", os.Getenv("LINE_MESSAGING_API_CHANNEL_ACCESS_TOKEN"), group, user)
		if memberStatus == 400 || memberStatus == 404 {
			writeAPIError(w, 403, "not_member", "無法確認你是此群組成員。")
			return
		}
		if memberStatus != 200 {
			writeAPIError(w, 503, "unavailable", "暫時無法確認群組身分。")
			return
		}
		id, err := store.ExchangeLaunch(ctx, input.Launch, user)
		if errors.Is(err, storage.ErrExpiredLaunch) {
			writeAPIError(w, 403, "expired_launch", "打卡連結已過期。")
			return
		}
		if err != nil {
			writeAPIError(w, 503, "unavailable", "暫時無法建立打卡操作。")
			return
		}
		writeJSON(w, 200, struct {
			RequestID string `json:"requestId"`
		}{id})
	})))
}

// A 200 from the Messaging API is the only positive membership evidence.
func checkGroupMember(ctx context.Context, client *http.Client, baseURL, token, group, user string) int {
	lookupCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	endpoint := baseURL + "/v2/bot/group/" + url.PathEscape(group) + "/member/" + url.PathEscape(user)
	req, err := http.NewRequestWithContext(lookupCtx, http.MethodGet, endpoint, nil)
	if err != nil {
		return 0
	}
	req.Header.Set("Authorization", "Bearer "+token)
	resp, err := client.Do(req)
	if err != nil {
		return 0
	}
	// Closing this read-only response cannot commit application state.
	defer func() { _ = resp.Body.Close() }()
	return resp.StatusCode
}
