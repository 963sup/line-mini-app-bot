package httpadapter

import (
	"encoding/json"
	"net/http"
	"regexp"
)

func writeJSON(w http.ResponseWriter, status int, value any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.Header().Set("Cache-Control", "no-store")
	w.Header().Set("X-Content-Type-Options", "nosniff")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(value)
}

func writeAPIError(w http.ResponseWriter, status int, code, message string) {
	type detail struct {
		Code    string `json:"code"`
		Message string `json:"message"`
	}
	writeJSON(w, status, struct {
		Error detail `json:"error"`
	}{Error: detail{Code: code, Message: message}})
}

func staticAssetHandler(content []byte, contentType string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", contentType)
		w.Header().Set("Cache-Control", "no-store")
		w.Header().Set("X-Content-Type-Options", "nosniff")
		w.Header().Set("Referrer-Policy", "no-referrer")
		w.Header().Set("Permissions-Policy", "geolocation=(self)")
		w.Header().Set("Content-Security-Policy",
			"default-src 'none'; script-src 'self'; style-src 'self'; connect-src 'self'; "+
				"base-uri 'none'; form-action 'none'; frame-ancestors 'none'")
		_, _ = w.Write(content)
	}
}

// Attendance request and summary job IDs use the same persisted SHA-256 format.
var requestIDPattern = regexp.MustCompile(`^[0-9a-f]{64}$`)
