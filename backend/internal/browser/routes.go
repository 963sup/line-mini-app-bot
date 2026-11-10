package browser

import (
	"net/http"
	"strings"
)

func serveMiniApp(w http.ResponseWriter, r *http.Request) {
	page := ""
	switch r.URL.Path {
	case "/miniapp":
		page = "miniapp.html"
		if r.URL.Query().Has("launch") || r.URL.Query().Has("liff.state") {
			page = "miniapp/punch.html"
		}
	case "/miniapp/history", "/miniapp/correction", "/miniapp/review", "/miniapp/punch":
		page = strings.TrimPrefix(r.URL.Path, "/") + ".html"
	default:
		http.NotFound(w, r)
		return
	}
	contents, err := ReadFile(page)
	if err != nil {
		http.Error(w, "無法載入頁面。", http.StatusServiceUnavailable)
		return
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.Header().Set("Cache-Control", "no-store")
	w.Header().Set("Referrer-Policy", "no-referrer")
	w.Header().Set("X-Content-Type-Options", "nosniff")
	w.Header().Set("Permissions-Policy", "geolocation=(self)")
	w.Header().Set("Content-Security-Policy", "default-src 'none'; script-src "+ScriptPolicy(contents)+"; style-src 'self'; connect-src 'self' https://api.line.me https://access.line.me https://liff.line.me https://liff.line-scdn.net; frame-src https://access.line.me; img-src 'self' data:; base-uri 'none'; form-action 'none'; frame-ancestors 'none'")
	_, _ = w.Write(contents)
}

func Register(mux *http.ServeMux) error {
	if err := Validate(); err != nil {
		return err
	}
	for _, route := range []string{"/miniapp", "/miniapp/history", "/miniapp/correction", "/miniapp/review", "/miniapp/punch"} {
		mux.HandleFunc("GET "+route, serveMiniApp)
	}
	mux.Handle("GET /assets/", Assets())
	mux.Handle("GET /_next/", Assets())
	return nil
}
