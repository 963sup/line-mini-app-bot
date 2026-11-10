package httpadapter

import (
	"fmt"
	"html/template"
	"net/http"
	"os"
	"regexp"
	"strings"
	"time"

	"github.com/963sup/line-mini-app-bot/internal/web"
)

var miniAppFiles = web.Files

var miniAppTemplates = template.Must(template.ParseFS(miniAppFiles,
	"pages/layout.html", "pages/home.html", "pages/history.html", "pages/correction.html", "pages/review.html"))

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

func RegisterMiniApp(mux *http.ServeMux) (string, error) {
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
	// Existing URLs remain aliases while pages use the descriptive asset names.
	assets := map[string]string{
		"miniapp.js": "static/js/session.js", "session.js": "static/js/session.js",
		"attendance.js": "static/js/punch.js", "punch.js": "static/js/punch.js",
		"app-api.js": "static/js/api.js", "api.js": "static/js/api.js",
		"time.js":    "static/js/time.js",
		"history.js": "static/js/history.js", "correction.js": "static/js/correction.js", "review.js": "static/js/review.js",
		"miniapp.css": "static/css/app.css", "app.css": "static/css/app.css",
	}
	for file, path := range assets {
		contents, err := miniAppFiles.ReadFile(path)
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
		contents, err := miniAppFiles.ReadFile("pages/punch.html")
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
