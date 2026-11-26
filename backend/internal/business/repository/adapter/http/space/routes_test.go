package spacehttp

import (
	repositoryApplicationDtoSpace "github.com/963sup/line-mini-app-bot/internal/business/repository/application/dto/space"
	repositoryApplicationSpace "github.com/963sup/line-mini-app-bot/internal/business/repository/application/space"
	repositoryDomainEntity "github.com/963sup/line-mini-app-bot/internal/business/repository/domain/entity"
	repositoryDomainEntitySpace "github.com/963sup/line-mini-app-bot/internal/business/repository/domain/entity/space"
)

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
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

func (s repos) Get(ctx context.Context, actor, id string) (repositoryDomainEntity.Repository, error) {
	if s.denied {
		return repositoryDomainEntity.Repository{}, repositoryDomainEntity.ErrNotFound
	}
	return repositoryDomainEntity.Repository{ID: id, OwnerType: repositoryDomainEntity.OwnerUser, OwnerID: "U" + strings.Repeat("b", 32), Status: "ACTIVE"}, nil
}

type store struct {
	calls            int
	actor, key, name string
}

func (s *store) List(context.Context, repositoryDomainEntity.Repository, string, string) (repositoryApplicationDtoSpace.Page, error) {
	s.calls++
	return repositoryApplicationDtoSpace.Page{Items: []repositoryDomainEntitySpace.Space{}}, nil
}

func (s *store) Get(context.Context, repositoryDomainEntity.Repository, string) (repositoryDomainEntitySpace.Space, error) {
	s.calls++
	return repositoryDomainEntitySpace.Space{}, repositoryDomainEntitySpace.ErrNotFound
}

func (s *store) Create(ctx context.Context, r repositoryDomainEntity.Repository, actor, key, name string) (repositoryDomainEntitySpace.Space, error) {
	s.calls++
	s.actor, s.key, s.name = actor, key, name
	return repositoryDomainEntitySpace.Space{ID: strings.Repeat("c", 32), RepositoryID: r.ID, Name: name, CreatorID: actor, Version: 1}, nil
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
			Register(mux, &repositoryApplicationSpace.Service{Repositories: repos{c.denied}, Store: st}, auth{c.unauth})
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

type unavailableRepositories struct{ err error }

func (r unavailableRepositories) Get(context.Context, string, string) (repositoryDomainEntity.Repository, error) {
	return repositoryDomainEntity.Repository{}, r.err
}

func TestRepositoryFailureDetails(t *testing.T) {
	failure := status.Error(codes.PermissionDenied, "repository read denied")
	mux := http.NewServeMux()
	Register(mux, &repositoryApplicationSpace.Service{Repositories: unavailableRepositories{failure}, Store: &store{}}, auth{})
	path := "/api/v1/repositories/" + strings.Repeat("a", 32) + "/spaces"
	for _, mode := range []string{"1", "0"} {
		t.Setenv("DEV_ERROR_DETAILS", mode)
		w := httptest.NewRecorder()
		mux.ServeHTTP(w, httptest.NewRequest("GET", path, nil))
		var result struct {
			Error struct{ Code, Message string }
		}
		if err := json.Unmarshal(w.Body.Bytes(), &result); err != nil {
			t.Fatal(err)
		}
		if w.Code != 503 || result.Error.Code != "unavailable" || (mode == "1" && result.Error.Message != failure.Error()) || (mode == "0" && strings.Contains(result.Error.Message, "repository read denied")) {
			t.Fatalf("mode %s: %d %s", mode, w.Code, w.Body.String())
		}
	}
}
