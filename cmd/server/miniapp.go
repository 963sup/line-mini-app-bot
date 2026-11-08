package main

import (
	"context"
	_ "embed"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"regexp"
	"strings"
	"time"
)

//go:embed web/miniapp.html
var miniAppPage []byte

//go:embed web/miniapp.js
var miniAppScript []byte

//go:embed web/attendance.js
var attendanceScript []byte

type miniApp struct {
	channelID string
	liffID    string
	client    *http.Client
}

func registerMiniApp(mux *http.ServeMux) (string, error) {
	app := &miniApp{
		channelID: os.Getenv("LINE_MINI_APP_CHANNEL_ID"),
		liffID:    os.Getenv("LINE_MINI_APP_LIFF_ID"),
		client:    &http.Client{Timeout: 5 * time.Second},
	}
	if err := app.validateConfig(); err != nil {
		return "", err
	}
	mux.HandleFunc("GET /miniapp", serveMiniApp)
	mux.HandleFunc("GET /assets/miniapp.js", staticAssetHandler(miniAppScript, "text/javascript; charset=utf-8"))
	mux.HandleFunc("GET /assets/attendance.js", staticAssetHandler(attendanceScript, "text/javascript; charset=utf-8"))
	mux.HandleFunc("GET /api/v1/miniapp/config", app.handleConfig)
	mux.HandleFunc("GET /api/v1/me", app.handleMe)
	if app.liffID == "" {
		return "", nil
	}
	return "https://miniapp.line.me/" + app.liffID, nil
}

func (app *miniApp) validateConfig() error {
	if app.channelID == "" && app.liffID == "" {
		return nil
	}
	if !regexp.MustCompile(`^[0-9]+$`).MatchString(app.channelID) {
		return fmt.Errorf("missing or invalid configuration: %s", "LINE_MINI_APP_CHANNEL_ID")
	}
	if !regexp.MustCompile(`^[0-9]+-[A-Za-z0-9]+$`).MatchString(app.liffID) {
		return fmt.Errorf("missing or invalid configuration: %s", "LINE_MINI_APP_LIFF_ID")
	}
	return nil
}

func (app *miniApp) handleConfig(w http.ResponseWriter, r *http.Request) {
	if app.liffID == "" {
		writeAPIError(w, http.StatusServiceUnavailable, "not_configured", "MINI App 尚未設定。")
		return
	}
	writeJSON(w, http.StatusOK, struct {
		LIFFID string `json:"liffId"`
	}{LIFFID: app.liffID})
}

func serveMiniApp(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.Header().Set("Cache-Control", "no-store")
	w.Header().Set("Referrer-Policy", "no-referrer")
	w.Header().Set("X-Content-Type-Options", "nosniff")
	w.Header().Set("Permissions-Policy", "geolocation=(self)")
	w.Header().Set("Content-Security-Policy",
		"default-src 'none'; script-src 'self' https://static.line-scdn.net; style-src 'self'; "+
			"connect-src 'self' https://api.line.me https://access.line.me https://liff.line.me https://liff.line-scdn.net; "+
			"frame-src https://access.line.me; base-uri 'none'; form-action 'none'; frame-ancestors 'none'")
	_, _ = w.Write(miniAppPage)
}

func (app *miniApp) handleMe(w http.ResponseWriter, r *http.Request) {
	user, name, ok := app.authenticate(w, r)
	if !ok {
		return
	}
	writeJSON(w, http.StatusOK, struct {
		UserID      string `json:"userId"`
		DisplayName string `json:"displayName"`
	}{UserID: user, DisplayName: name})
}

func (app *miniApp) authenticate(w http.ResponseWriter, r *http.Request) (string, string, bool) {
	if app.channelID == "" {
		writeAPIError(w, http.StatusServiceUnavailable, "not_configured", "MINI App 尚未設定。")
		return "", "", false
	}
	authorization := strings.Fields(r.Header.Get("Authorization"))
	if len(authorization) != 2 || !strings.EqualFold(authorization[0], "Bearer") {
		writeAPIError(w, http.StatusUnauthorized, "unauthenticated", "請先登入 LINE。")
		return "", "", false
	}
	if len(authorization[1]) > 16384 {
		writeAPIError(w, http.StatusUnauthorized, "unauthenticated", "無效的 LINE token。")
		return "", "", false
	}
	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()
	form := url.Values{"id_token": {authorization[1]}, "client_id": {app.channelID}}
	req, err := http.NewRequestWithContext(
		ctx,
		http.MethodPost,
		"https://api.line.me/oauth2/v2.1/verify",
		strings.NewReader(form.Encode()),
	)
	if err != nil {
		writeAPIError(w, http.StatusServiceUnavailable, "unavailable", "暫時無法驗證 LINE 身分。")
		return "", "", false
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	response, err := app.client.Do(req)
	if err != nil {
		writeAPIError(w, http.StatusServiceUnavailable, "unavailable", "暫時無法驗證 LINE 身分。")
		return "", "", false
	}
	// The response is read-only; closing cannot commit application state.
	defer func() { _ = response.Body.Close() }()
	if response.StatusCode == http.StatusBadRequest || response.StatusCode == http.StatusUnauthorized {
		writeAPIError(w, http.StatusUnauthorized, "unauthenticated", "LINE 登入已失效，請重新登入。")
		return "", "", false
	}
	if response.StatusCode != http.StatusOK {
		writeAPIError(w, http.StatusServiceUnavailable, "unavailable", "暫時無法驗證 LINE 身分。")
		return "", "", false
	}
	var claims struct {
		Issuer   string `json:"iss"`
		Subject  string `json:"sub"`
		Audience string `json:"aud"`
		Expires  int64  `json:"exp"`
		Name     string `json:"name"`
	}
	if err := json.NewDecoder(io.LimitReader(response.Body, 16<<10)).Decode(&claims); err != nil {
		writeAPIError(w, http.StatusServiceUnavailable, "unavailable", "LINE 驗證回應無效。")
		return "", "", false
	}
	validChannel := claims.Issuer == "https://access.line.me" && claims.Audience == app.channelID
	validUser := claims.Subject != "" && claims.Expires > time.Now().Unix()
	if !validChannel || !validUser {
		writeAPIError(w, http.StatusUnauthorized, "unauthenticated", "LINE 身分驗證未通過。")
		return "", "", false
	}
	return claims.Subject, claims.Name, true
}
