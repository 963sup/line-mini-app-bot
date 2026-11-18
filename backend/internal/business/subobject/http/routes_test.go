package subobjecthttp

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/963sup/line-mini-app-bot/internal/business/repository"
	feature "github.com/963sup/line-mini-app-bot/internal/business/subobject"
	"github.com/963sup/line-mini-app-bot/internal/httpapi/generated"
)

type auth struct{ deny bool }

func (a auth) Authenticate(w http.ResponseWriter, _ *http.Request) (string, string, bool) {
	if a.deny {
		w.WriteHeader(401)
		return "", "", false
	}
	return "owner", "Owner", true
}

type repos struct{ deny bool }

func (r repos) Get(_ context.Context, _, id string) (repository.Repository, error) {
	if r.deny {
		return repository.Repository{}, repository.ErrNotFound
	}
	return repository.Repository{ID: id, GroupID: "C" + strings.Repeat("b", 32), Status: "ACTIVE"}, nil
}

type memory struct {
	child feature.Subobject
	calls int
	pid   string
}

func (m *memory) List(_ context.Context, _ repository.Repository, _, pid, _ string) (feature.Page, error) {
	m.calls++
	m.pid = pid
	return feature.Page{Items: []feature.Subobject{m.child}}, nil
}

func (m *memory) Get(_ context.Context, _ repository.Repository, _, pid, _ string) (feature.Subobject, error) {
	m.calls++
	m.pid = pid
	return m.child, nil
}

func (m *memory) Create(_ context.Context, r repository.Repository, actor, pid, _ string, c feature.Content) (feature.Subobject, error) {
	m.calls++
	m.pid = pid
	status, e := feature.Status(c)
	if e != nil {
		return feature.Subobject{}, e
	}
	m.child = feature.Subobject{Content: c, ID: strings.Repeat("c", 32), RepositoryID: r.ID, ParentObjectID: pid, CreatorID: actor, Status: status, Version: 1, CreatedAt: time.Now().UTC(), UpdatedAt: time.Now().UTC()}
	return m.child, nil
}

func (m *memory) Update(_ context.Context, _ repository.Repository, actor, pid, _ string, input feature.Update) (feature.Subobject, error) {
	m.calls++
	m.pid = pid
	next, e := feature.ApplyUpdate(m.child, actor, input)
	if e == nil {
		m.child = next
	}
	return next, e
}

func TestHTTPSubobjectContractAndIndependentScope(t *testing.T) {
	rid, pid := strings.Repeat("a", 32), strings.Repeat("b", 32)
	path := "/api/v1/repositories/" + rid + "/objects/" + pid + "/subobjects"
	base := `{"name":"Lay","type":"work","quantity":"140.000001","unit":"M"}`
	m := &memory{}
	mux := http.NewServeMux()
	Register(mux, &feature.Service{Repositories: repos{}, Store: m}, auth{})
	request := func(method, path, body, key string) *httptest.ResponseRecorder {
		r := httptest.NewRequest(method, path, strings.NewReader(body))
		r.Header.Set("Content-Type", "application/json")
		if key != "" {
			r.Header.Set("Idempotency-Key", key)
		}
		w := httptest.NewRecorder()
		mux.ServeHTTP(w, r)
		return w
	}
	for _, body := range []string{strings.TrimSuffix(base, "}") + `,"parentObjectId":"` + pid + `"}`, strings.TrimSuffix(base, "}") + `,"sourceDocument":"private"}`, strings.Replace(base, `"type":"work"`, `"type":"Engineering"`, 1)} {
		w := request("POST", path, body, "subobject-test-key")
		if w.Code != 400 || m.calls != 0 {
			t.Fatal(w.Code, w.Body.String())
		}
	}
	w := request("POST", path, base, "subobject-test-key")
	if w.Code != 200 || m.pid != pid {
		t.Fatal(w.Code, w.Body.String())
	}
	var j generated.Subobject
	if json.Unmarshal(w.Body.Bytes(), &j) != nil || j.CompletedQuantity == nil || *j.CompletedQuantity != "0" || j.ParentObjectId != pid {
		t.Fatal(w.Body.String())
	}
	childPath := path + "/" + j.Id
	patch := strings.TrimSuffix(base, "}") + `,"completedQuantity":"40.000001","expectedVersion":1}`
	w = request("PATCH", childPath, patch, "")
	if w.Code != 200 || m.child.Status != "in_progress" {
		t.Fatal(w.Code, w.Body.String())
	}
	w = request("PATCH", childPath, patch, "")
	if w.Code != 409 {
		t.Fatal(w.Code, w.Body.String())
	}
	for _, p := range []string{path, childPath} {
		w = request("GET", p, "", "")
		if w.Code != 200 || w.Header().Get("Cache-Control") != "no-store" {
			t.Fatal(w.Code)
		}
	}
	if w = request("POST", path, base, ""); w.Code != 400 {
		t.Fatal("missing key", w.Code)
	}
	if w = request("POST", childPath+"/subobjects", base, "another-test-key"); w.Code != 404 {
		t.Fatal("grandchild route", w.Code)
	}
	for _, deny := range []bool{true, false} {
		denied := http.NewServeMux()
		Register(denied, &feature.Service{Repositories: repos{deny: true}, Store: &memory{}}, auth{deny: deny})
		req := httptest.NewRequest("GET", path, nil)
		out := httptest.NewRecorder()
		denied.ServeHTTP(out, req)
		expected := 404
		if deny {
			expected = 401
		}
		if out.Code != expected {
			t.Fatal("authentication/scope", out.Code)
		}
	}
}
