package httpadapter

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestSummaryWorkerRequiresAuthentication(t *testing.T) {
	r := httptest.NewRequest(http.MethodPost, "/internal/summary", strings.NewReader(`{"group":"g","id":"job"}`))
	w := httptest.NewRecorder()
	SummaryWorker(nil, "https://example.test/internal/summary", "tasks@example.test")(w, r)
	if w.Code != http.StatusUnauthorized {
		t.Fatalf("status=%d", w.Code)
	}
}
