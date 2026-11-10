package main

import (
	"context"
	"embed"
	"encoding/json"
	"fmt"
	"html/template"
	"io"
	"net/http"
	"net/url"
	"os"
	"regexp"
	"strings"
	"time"
)

//go:embed web/*.html web/*.js web/miniapp.css
var miniAppFiles embed.FS

var miniAppTemplates = template.Must(template.ParseFS(miniAppFiles,
	"web/miniapp.html", "web/home.html", "web/history.html", "web/correction.html", "web/review.html"))

var miniAppRoutes = map[string]struct{ Page, Title string }{
	"/miniapp":            {"home", "出勤服務"},
	"/miniapp/history":    {"history", "本人月紀錄"},
	"/miniapp/correction": {"correction", "更正與申請結果"},
	"/miniapp/review":     {"review", "授權審核"},
}

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
	for route := range miniAppRoutes {
		mux.HandleFunc("GET "+route, serveMiniApp)
	}
	for _, file := range []string{"miniapp.js", "attendance.js", "app-api.js", "history.js", "correction.js", "review.js", "miniapp.css"} {
		contents, err := miniAppFiles.ReadFile("web/" + file)
		if err != nil {
			return "", err
		}
		contentType := "text/javascript; charset=utf-8"
		if strings.HasSuffix(file, ".css") {
			contentType = "text/css; charset=utf-8"
		}
		mux.HandleFunc("GET /assets/"+file, staticAssetHandler(contents, contentType))
	}
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
	page, ok := miniAppRoutes[r.URL.Path]
	if !ok {
		http.NotFound(w, r)
		return
	}
	if r.URL.Path == "/miniapp" && (r.URL.Query().Has("launch") || r.URL.Query().Has("liff.state")) {
		contents, err := miniAppFiles.ReadFile("web/attendance.html")
		if err != nil {
			http.Error(w, "無法載入打卡頁面。", http.StatusInternalServerError)
			return
		}
		_, _ = w.Write(contents)
		return
	}
	if err := miniAppTemplates.ExecuteTemplate(w, "miniapp", page); err != nil {
		return
	}
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
