package spacehttp

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/963sup/line-mini-app-bot/internal/business/repository"
	feature "github.com/963sup/line-mini-app-bot/internal/business/repository/space"
)

type auth struct{ denied bool }

func (a auth) Authenticate(w http.ResponseWriter, r *http.Request) (string, string, bool) {
	if a.denied {
		w.WriteHeader(401)
		return "", "", false
	}
	return "trusted-actor", "Actor", true
}

type repos struct{ denied bool }

func (s repos) Get(ctx context.Context, actor, id string) (repository.Repository, error) {
	if s.denied {
		return repository.Repository{}, repository.ErrNotFound
	}
	return repository.Repository{ID: id, GroupID: "C" + strings.Repeat("b", 32), Status: "ACTIVE"}, nil
}

type store struct {
	calls            int
	actor, key, name string
}

func (s *store) List(context.Context, repository.Repository, string, string) (feature.Page, error) {
	s.calls++
	return feature.Page{Items: []feature.Space{}}, nil
}

func (s *store) Get(context.Context, repository.Repository, string) (feature.Space, error) {
	s.calls++
	return feature.Space{}, feature.ErrNotFound
}

func (s *store) Create(ctx context.Context, r repository.Repository, actor, key, name string) (feature.Space, error) {
	s.calls++
	s.actor, s.key, s.name = actor, key, name
	return feature.Space{ID: strings.Repeat("c", 32), RepositoryID: r.ID, Name: name, CreatorID: actor, Version: 1}, nil
}

func TestHTTPAuthorityAndIDBoundary(t *testing.T) {
	path := "/api/v1/repositories/" + strings.Repeat("a", 32) + "/spaces"
	for _, c := range []struct {
		name, method, path, body string
		status                   int
		unauth, denied           bool
	}{
		{"unauthenticated", "GET", path, "", 401, true, false},
		{"current membership denied", "GET", path, "", 404, false, true},
		{"list", "GET", path, "", 200, false, false},
		{"create", "POST", path, `{"name":" 一樓 "}`, 200, false, false},
		{"client actor not trusted", "POST", path, `{"name":"x","userId":"attacker"}`, 400, false, false},
		{"group not accepted", "POST", path, `{"name":"x","groupId":"C00000000000000000000000000000000"}`, 400, false, false},
		{"group ID not business ID", "GET", path + "/C" + strings.Repeat("c", 32), "", 400, false, false},
		{"empty name", "POST", path, `{"name":" "}`, 400, false, false},
	} {
		t.Run(c.name, func(t *testing.T) {
			st := &store{}
			mux := http.NewServeMux()
			Register(mux, &feature.Service{Repositories: repos{c.denied}, Store: st}, auth{c.unauth})
			r := httptest.NewRequest(c.method, c.path, strings.NewReader(c.body))
			r.Header.Set("Content-Type", "application/json")
			r.Header.Set("Idempotency-Key", "space-intent-123456")
			w := httptest.NewRecorder()
			mux.ServeHTTP(w, r)
			if w.Code != c.status {
				t.Fatalf("status %d want %d: %s", w.Code, c.status, w.Body.String())
			}
			if c.status >= 400 && st.calls != 0 {
				t.Fatal("denied request reached storage")
			}
			if c.name == "create" && (st.actor != "trusted-actor" || st.name != "一樓" || st.key != "space-intent-123456") {
				t.Fatal("authority or normalized intent lost")
			}
		})
	}
}
