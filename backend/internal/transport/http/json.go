package httptransport

import (
	"encoding/json"
	"net/http"
)

func WriteJSON(w http.ResponseWriter, status int, value any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.Header().Set("Cache-Control", "no-store")
	w.Header().Set("X-Content-Type-Options", "nosniff")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(value)
}

func WriteAPIError(w http.ResponseWriter, status int, code, message string) {
	type detail struct {
		Code    string `json:"code"`
		Message string `json:"message"`
	}
	WriteJSON(w, status, struct {
		Error detail `json:"error"`
	}{Error: detail{Code: code, Message: message}})
}

func ValidateUniqueQuery(w http.ResponseWriter, r *http.Request) bool {
	for _, values := range r.URL.Query() {
		if len(values) > 1 {
			WriteAPIError(w, http.StatusBadRequest, "invalid_input", "請勿重複查詢參數。")
			return false
		}
	}
	return true
}
