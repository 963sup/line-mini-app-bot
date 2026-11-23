package httptransport

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestServiceFailureDiagnostics(t *testing.T) {
	var logs bytes.Buffer
	previous := slog.Default()
	slog.SetDefault(slog.New(slog.NewJSONHandler(&logs, nil)))
	t.Cleanup(func() { slog.SetDefault(previous) })
	t.Setenv("DEV_ERROR_DETAILS", "")
	secret := "private-token-and-request-body"
	for _, resource := range []bool{false, true} {
		logs.Reset()
		handler := RequestDiagnostics(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			err := fmt.Errorf("private wrapper %s: %w", secret, errors.New(secret))
			if resource {
				WriteResourceError(w, 503, "unavailable", "Try again", err, true)
			} else {
				WriteServiceError(w, 503, "unavailable", "Try again", err)
			}
		}))
		request := httptest.NewRequest("GET", "/private?token="+secret, nil)
		request.Header.Set("X-Request-ID", secret)
		w := httptest.NewRecorder()
		handler.ServeHTTP(w, request)
		id := w.Header().Get("X-Request-ID")
		var entry struct {
			RequestID  string   `json:"request_id"`
			Status     int      `json:"status"`
			Code       string   `json:"code"`
			ErrorTypes []string `json:"error_types"`
		}
		if err := json.Unmarshal(logs.Bytes(), &entry); err != nil {
			t.Fatal(err)
		}
		if id == "" || id == secret || entry.RequestID != id || entry.Status != 503 || entry.Code != "unavailable" || len(entry.ErrorTypes) != 2 {
			t.Fatalf("unexpected diagnostic metadata: %+v", entry)
		}
		if strings.Contains(logs.String()+w.Body.String(), secret) {
			t.Fatal("diagnostics exposed private request or error content")
		}
		if w.Code != 503 || !strings.Contains(w.Body.String(), `"message":"Try again"`) || resource && !strings.Contains(w.Body.String(), `"retrySameRequest":true`) {
			t.Fatal("changed public failure semantics", w.Code, w.Body.String())
		}
	}
	logs.Reset()
	WriteServiceError(httptest.NewRecorder(), 404, "not_found", "Missing", errors.New(secret))
	if logs.Len() != 0 {
		t.Fatal("expected client errors must not produce server failure logs")
	}
}
