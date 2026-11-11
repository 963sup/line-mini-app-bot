package summaryhttp

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestSummaryWorkerRequiresAuthentication(t *testing.T) {
	r := httptest.NewRequest(http.MethodPost, "/internal/summary", strings.NewReader(`{"group":"g","id":"job"}`))
	w := httptest.NewRecorder()
	SummaryWorker(nil, nil)(w, r)
	if w.Code != http.StatusUnauthorized {
		t.Fatalf("status=%d", w.Code)
	}
}
