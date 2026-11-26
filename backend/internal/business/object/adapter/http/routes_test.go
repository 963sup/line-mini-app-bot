package objecthttp

import (
	repositoryDomainEntity "github.com/963sup/line-mini-app-bot/internal/business/repository/domain/entity"
)

import (
	objectApplication "github.com/963sup/line-mini-app-bot/internal/business/object/application"
	objectApplicationDto "github.com/963sup/line-mini-app-bot/internal/business/object/application/dto"
	objectDomainEntity "github.com/963sup/line-mini-app-bot/internal/business/object/domain/entity"
)

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
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

func (r repositories) Get(_ context.Context, _, id string) (repositoryDomainEntity.Repository, error) {
	if r.deny {
		return repositoryDomainEntity.Repository{}, repositoryDomainEntity.ErrNotFound
	}
	return repositoryDomainEntity.Repository{ID: id, OwnerType: repositoryDomainEntity.OwnerUser, OwnerID: "U" + strings.Repeat("b", 32), Status: "ACTIVE"}, nil
}

type store struct {
	calls  int
	input  objectDomainEntity.Content
	update objectDomainEntity.Update
}

func (s *store) List(context.Context, repositoryDomainEntity.Repository, string, string, ...string) (objectApplicationDto.Page, error) {
	s.calls++
	return objectApplicationDto.Page{Items: []objectDomainEntity.Object{}}, nil
}

func (s *store) Get(context.Context, repositoryDomainEntity.Repository, string) (objectDomainEntity.Object, error) {
	s.calls++
	return objectDomainEntity.Object{}, objectDomainEntity.ErrNotFound
}

func (s *store) Create(_ context.Context, r repositoryDomainEntity.Repository, actor, _ string, c objectDomainEntity.Content) (objectDomainEntity.Object, error) {
	s.calls++
	s.input = c
	return objectDomainEntity.Object{Content: c, RepositoryID: r.ID, CreatorID: actor, Status: "not_started", Version: 1}, nil
}

func (s *store) Update(_ context.Context, _ repositoryDomainEntity.Repository, _, _ string, u objectDomainEntity.Update) (objectDomainEntity.Object, error) {
	s.calls++
	s.update = u
	return objectDomainEntity.Object{}, objectDomainEntity.ErrConflict
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
		{name: "create", method: "POST", path: path, body: `{"name":"項目"}`, status: 200},
		{name: "work initialized", method: "POST", path: path, body: strings.Replace(base, `"type":"工程"`, `"type":"work"`, 1), status: 200},
		{name: "negative completed", method: "POST", path: path, body: strings.TrimSuffix(strings.Replace(base, `"type":"工程"`, `"type":"work"`, 1), "}") + `,"completedQuantity":"-1"}`, status: 400},
		{name: "optional status", method: "PATCH", path: path + "/" + rid, body: strings.TrimSuffix(base, "}") + `,"completedQuantity":"0","expectedVersion":1}`, status: 409},
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
			Register(mux, &objectApplication.Service{Repositories: repositories{deny: c.denied}, Store: s}, auth{deny: c.noAuth})
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
			if c.name == "create" && (s.input.Quantity != "" || s.input.Type != "" || s.input.Unit != "" || s.input.SpaceID != "" || s.input.CompletedQuantity != "" || s.input.Name != "項目") {
				t.Fatalf("not normalized: %+v", s.input)
			}
			if c.name == "work initialized" && s.input.CompletedQuantity != "0" {
				t.Fatal("work not initialized")
			}
			if c.name == "optional status" && (s.update.Status != "" || s.update.CompletedQuantity == nil || *s.update.CompletedQuantity != "0") {
				t.Fatal("optional field lost")
			}
			if c.name == "version conflict" && s.update.Sources.SourceDocument != nil {
				t.Fatal("omission became clear")
			}
		})
	}
}

func TestHTTPSpaceAssignmentOmissionAndClear(t *testing.T) {
	rid := strings.Repeat("a", 32)
	sid := strings.Repeat("d", 32)
	for _, tc := range []struct {
		name, method, space string
		want                *string
	}{
		{name: "create assigned", method: "POST", space: `,"spaceId":"` + sid + `"`, want: &sid},
		{name: "patch omission", method: "PATCH"},
		{name: "patch clear", method: "PATCH", space: `,"spaceId":""`, want: new(string)},
	} {
		t.Run(tc.name, func(t *testing.T) {
			st := &store{}
			mux := http.NewServeMux()
			Register(mux, &objectApplication.Service{Repositories: repositories{}, Store: st}, auth{})
			path := "/api/v1/repositories/" + rid + "/objects"
			extra := ""
			wantStatus := 200
			if tc.method == "PATCH" {
				path += "/" + rid
				extra = `,"expectedVersion":1`
				wantStatus = 409
			}
			payload := `{"name":"物件","description":"","type":"work","quantity":"1","unit":"SET"` + extra + tc.space + `}`
			r := httptest.NewRequest(tc.method, path, strings.NewReader(payload))
			r.Header.Set("Content-Type", "application/json")
			r.Header.Set("Idempotency-Key", strings.Repeat("k", 16))
			w := httptest.NewRecorder()
			mux.ServeHTTP(w, r)
			if w.Code != wantStatus {
				t.Fatal(w.Code, w.Body.String())
			}
			if tc.method == "POST" {
				if st.input.SpaceID != sid {
					t.Fatal("create assignment lost")
				}
			} else if (tc.want == nil) != (st.update.SpaceID == nil) || (tc.want != nil && *tc.want != *st.update.SpaceID) {
				t.Fatal("patch omission/clear mapping lost")
			}
		})
	}
}

type candidateProbe struct{ calls int }

func (p *candidateProbe) PrepareCandidates(_ context.Context, actor, repository string, d objectApplicationDto.Document) (objectApplicationDto.CandidateResult, error) {
	if actor == "" || repository == "" {
		panic("lost authorized scope")
	}
	p.calls++
	return objectApplicationDto.CandidateResult{Candidates: []objectApplicationDto.Candidate{{Name: d.Filename, Quote: string(d.Data)}}}, nil
}

func TestCandidateRouteChecksScopeAndDoesNotCreate(t *testing.T) {
	for _, deny := range []bool{true, false} {
		probe := &candidateProbe{}
		st := &store{}
		mux := http.NewServeMux()
		Register(mux, &objectApplication.Service{Repositories: repositories{deny: deny}, Store: st, CandidatePreparer: probe}, auth{})
		r := httptest.NewRequest("POST", "/api/v1/repositories/"+strings.Repeat("a", 32)+"/objects/candidates", strings.NewReader(`{"filename":"test.txt","data":"YWJj"}`))
		r.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		mux.ServeHTTP(w, r)
		if deny {
			if w.Code != 404 || probe.calls != 0 {
				t.Fatalf("denied: %d %s", w.Code, w.Body.String())
			}
		} else if w.Code != 200 || probe.calls != 1 || !strings.Contains(w.Body.String(), `"quote":"abc"`) {
			t.Fatalf("extraction: %d %s", w.Code, w.Body.String())
		}
		if st.calls != 0 {
			t.Fatal("extraction touched object store")
		}
	}
}
