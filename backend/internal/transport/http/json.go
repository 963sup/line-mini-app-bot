package httptransport

import (
	"encoding/json"
	"io"
	"net/http"
	"os"
)

// ReadJSON accepts one JSON value and rejects unknown fields.
func ReadJSON(w http.ResponseWriter, r *http.Request, target any, invalidMessage string) bool {
	d := json.NewDecoder(r.Body)
	d.DisallowUnknownFields()
	if d.Decode(target) != nil || d.Decode(new(any)) != io.EOF {
		WriteAPIError(w, http.StatusBadRequest, "invalid_input", invalidMessage)
		return false
	}
	return true
}

func WriteJSON(w http.ResponseWriter, status int, value any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.Header().Set("Cache-Control", "no-store")
	w.Header().Set("X-Content-Type-Options", "nosniff")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(value)
}

func WriteAPIError(w http.ResponseWriter, status int, code, message string) {
	writeError(w, status, code, message, nil)
}

// WriteRequestError adds recovery guidance for this attempt only.
func WriteRequestError(w http.ResponseWriter, status int, code, message string, retrySameRequest bool) {
	writeError(w, status, code, message, &retrySameRequest)
}

func writeError(w http.ResponseWriter, status int, code, message string, retrySameRequest *bool) {
	type detail struct {
		RetrySameRequest *bool  `json:"retrySameRequest,omitempty"`
		Code             string `json:"code"`
		Message          string `json:"message"`
	}
	WriteJSON(w, status, struct {
		Error detail `json:"error"`
	}{Error: detail{Code: code, Message: message, RetrySameRequest: retrySameRequest}})
}

// WriteServiceError exposes the service message when running the local dev task.
func WriteServiceError(w http.ResponseWriter, status int, code, message string, err error) {
	if os.Getenv("DEV_ERROR_DETAILS") == "1" {
		message = err.Error()
	}
	WriteAPIError(w, status, code, message)
}

func WriteResourceError(w http.ResponseWriter, status int, code, message string, err error, retrySameRequest bool) {
	if os.Getenv("DEV_ERROR_DETAILS") == "1" {
		message = err.Error()
	}
	WriteRequestError(w, status, code, message, retrySameRequest)
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
