package journalhttp

import (
	"context"
	"github.com/963sup/line-mini-app-bot/internal/business/journal"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

type testAuth struct{}

func (testAuth) Authenticate(w http.ResponseWriter, r *http.Request) (string, string, bool) {
	if r.Header.Get("Authorization") != "Bearer test" {
		w.WriteHeader(401)
		return "", "", false
	}
	return "actor", "", true
}

type fixture struct {
	saved  journal.Entry
	filter journal.Filter
	key    string
}

func (f *fixture) Create(_ context.Context, actor, key string, d journal.Draft) (journal.Entry, error) {
	f.key = key
	f.saved = journal.Entry{Draft: d, ID: strings.Repeat("a", 32), CreatorID: actor, CreatedAt: d.WorkDate}
	return f.saved, nil
}
func (f *fixture) List(_ context.Context, actor string, filter journal.Filter) (journal.Page, error) {
	f.filter = filter
	return journal.Page{Items: []journal.Entry{f.saved}}, nil
}
func (f *fixture) Get(_ context.Context, actor, id string) (journal.Entry, error) {
	if id != f.saved.ID || actor != f.saved.CreatorID {
		return journal.Entry{}, journal.ErrNotFound
	}
	return f.saved, nil
}
func TestJournalRoutes(t *testing.T) {
	const base = "/api/v1/journals"
	const group = "C11111111111111111111111111111111"
	const key = "sample-journal-001"
	const draft = `{"spaceId":"` + group + `","workDate":"2026-11-18","personnel":{"management":[],"control":[],"workers":["甲"]},"todayItems":[{"name":"鋪設","completionPercent":40,"reference":{"type":"project","id":"aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"}},{"name":"整理","completionPercent":100}],"tomorrowItems":[]}`
	f := new(fixture)
	mux := http.NewServeMux()
	Register(mux, f, testAuth{})
	call := func(method, path, body string, auth bool, want int) string {
		t.Helper()
		r := httptest.NewRequest(method, path, strings.NewReader(body))
		if auth {
			r.Header.Set("Authorization", "Bearer test")
		}
		if method == "POST" {
			r.Header.Set("Content-Type", "application/json")
			r.Header.Set("Idempotency-Key", key)
		}
		w := httptest.NewRecorder()
		mux.ServeHTTP(w, r)
		if w.Code != want {
			t.Fatalf("%s %s: %d %s", method, path, w.Code, w.Body.String())
		}
		return w.Body.String()
	}
	call("GET", base, "", false, 401)
	call("POST", base, strings.Replace(draft, "100", "101", 1), true, 400)
	call("POST", base, draft, true, 201)
	if f.key != key || f.saved.CreatorID != "actor" || f.saved.Today[0].Reference.Kind != "project" || f.saved.Today[1].Reference != nil {
		t.Fatal("creation mapping")
	}
	list := call("GET", base+"?workDate=2026-11-18&spaceId="+group, "", true, 200)
	if f.filter.SpaceID != group || f.filter.WorkDate != "2026-11-18" || !strings.Contains(list, f.saved.ID) {
		t.Fatal("dashboard filter or response")
	}
	call("GET", base+"/"+f.saved.ID, "", true, 200)
	call("GET", base+"/"+strings.Repeat("b", 32), "", true, 404)
	mux = http.NewServeMux()
	Register(mux, nil, testAuth{})
	call("POST", base, draft, true, 503)
}
