package projecthttp

import (
	"encoding/json"
	"net/http/httptest"
	"strings"
	"testing"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func TestStorageFailureResponse(t *testing.T) {
	t.Setenv("DEV_ERROR_DETAILS", "")
	w := httptest.NewRecorder()
	writeError(w, status.Error(codes.PermissionDenied, "private storage details"))
	var response struct{ Error struct{ Code string } }
	if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
		t.Fatal(err)
	}
	if w.Code != 503 || response.Error.Code != "unavailable" || strings.Contains(w.Body.String(), "private storage details") {
		t.Fatalf("unexpected response: %d %s", w.Code, w.Body.String())
	}
}

func TestDevelopmentStorageFailureResponse(t *testing.T) {
	t.Setenv("DEV_ERROR_DETAILS", "1")
	failure := status.Error(codes.PermissionDenied, "storage access denied")
	w := httptest.NewRecorder()
	writeError(w, failure)
	var response struct {
		Error struct{ Code, Message string }
	}
	if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
		t.Fatal(err)
	}
	if w.Code != 503 || response.Error.Code != "unavailable" || response.Error.Message != failure.Error() {
		t.Fatal(w.Body.String())
	}
}
