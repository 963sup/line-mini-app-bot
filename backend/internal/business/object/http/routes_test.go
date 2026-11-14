package objecthttp

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/963sup/line-mini-app-bot/internal/business/object"
	"github.com/963sup/line-mini-app-bot/internal/business/repository"
)

type auth struct{ deny bool }

func (a auth) Authenticate(w http.ResponseWriter, _ *http.Request) (string, string, bool) {
	if a.deny {
		w.WriteHeader(401)
		return "", "", false
	}
	return "A", "Actor", true
}

type repositories struct{ deny bool }

func (r repositories) Get(_ context.Context, _, id string) (repository.Repository, error) {
	if r.deny {
		return repository.Repository{}, repository.ErrNotFound
	}
	return repository.Repository{ID: id, GroupID: "C" + strings.Repeat("b", 32), Status: "ACTIVE"}, nil
}

type store struct {
	calls  int
	input  object.Content
	update object.Update
}

func (s *store) List(context.Context, repository.Repository, string, string) (object.Page, error) {
	s.calls++
	return object.Page{Items: []object.Object{}}, nil
}

func (s *store) Get(context.Context, repository.Repository, string) (object.Object, error) {
	s.calls++
	return object.Object{}, object.ErrNotFound
}

func (s *store) Create(_ context.Context, r repository.Repository, actor, _ string, c object.Content) (object.Object, error) {
	s.calls++
	s.input = c
	return object.Object{Content: c, RepositoryID: r.ID, CreatorID: actor, Status: "not_started", Version: 1}, nil
}

func (s *store) Update(_ context.Context, _ repository.Repository, _, _ string, u object.Update) (object.Object, error) {
	s.calls++
	s.update = u
	return object.Object{}, object.ErrConflict
}

func TestHTTPBoundariesAndSourceOmission(t *testing.T) {
	rid := strings.Repeat("a", 32)
	path := "/api/v1/repositories/" + rid + "/objects"
	base := `{"name":"項目","description":"說明","type":"工程","quantity":"11.00","unit":"SET"}`
	cases := []struct {
		name, method, path, body string
		status                   int
		noAuth, denied           bool
	}{
		{name: "unauthenticated", method: "GET", path: path, status: 401, noAuth: true},
		{name: "cross repository denied", method: "GET", path: path, status: 404, denied: true},
		{name: "list", method: "GET", path: path, status: 200},
		{name: "create", method: "POST", path: path, body: base, status: 200},
		{name: "zero", method: "POST", path: path, body: strings.Replace(base, "11.00", "0", 1), status: 400},
		{name: "unknown actor", method: "POST", path: path, body: strings.TrimSuffix(base, "}") + `,"creatorId":"B"}`, status: 400},
		{name: "trailing json", method: "POST", path: path, body: base + ` {}`, status: 400},
		{name: "bad enum", method: "PATCH", path: path + "/" + rid, body: strings.TrimSuffix(base, "}") + `,"status":"approved","expectedVersion":1}`, status: 400},
		{name: "version conflict", method: "PATCH", path: path + "/" + rid, body: strings.TrimSuffix(base, "}") + `,"status":"completed","expectedVersion":1}`, status: 409},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			s := &store{}
			mux := http.NewServeMux()
			Register(mux, &object.Service{Repositories: repositories{deny: c.denied}, Store: s}, auth{deny: c.noAuth})
			r := httptest.NewRequest(c.method, c.path, strings.NewReader(c.body))
			r.Header.Set("Content-Type", "application/json")
			r.Header.Set("Idempotency-Key", strings.Repeat("k", 16))
			w := httptest.NewRecorder()
			mux.ServeHTTP(w, r)
			if w.Code != c.status {
				t.Fatalf("got %d want %d: %s", w.Code, c.status, w.Body.String())
			}
			if w.Header().Get("Cache-Control") != "no-store" {
				t.Fatal("missing no-store")
			}
			if c.status != 200 && c.status != 409 && s.calls != 0 {
				t.Fatal("rejected request reached store")
			}
			if c.name == "create" && (s.input.Quantity != "11" || s.input.Name != "項目") {
				t.Fatalf("not normalized: %+v", s.input)
			}
			if c.name == "version conflict" && s.update.Sources.SourceDocument != nil {
				t.Fatal("omission became clear")
			}
		})
	}
}
