package milestonehttp

import (
	"encoding/json"
	"net/http/httptest"
	"strings"
	"testing"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func TestStorageFailureResponse(t *testing.T) {
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
