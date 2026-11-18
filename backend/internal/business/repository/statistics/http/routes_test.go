package statisticshttp

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

type authenticated struct{}

func (authenticated) Authenticate(http.ResponseWriter, *http.Request) (string, string, bool) {
	return "actor", "Actor", true
}

func TestStatisticsRouteMethods(t *testing.T) {
	mux := http.NewServeMux()
	Register(mux, nil, authenticated{})
	path := "/api/v1/repositories/" + strings.Repeat("a", 32) + "/statistics"
	// Reads reach the unconfigured service. Writes are rejected by the mux.
	for method, want := range map[string]int{"GET": 503, "HEAD": 503, "POST": 405, "PATCH": 405} {
		w := httptest.NewRecorder()
		mux.ServeHTTP(w, httptest.NewRequest(method, path, nil))
		if w.Code != want {
			t.Errorf("%s: got %d, want %d", method, w.Code, want)
		}
	}
}
