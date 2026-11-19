package projecthttp

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	groupdomain "github.com/963sup/line-mini-app-bot/internal/business/group"
	feature "github.com/963sup/line-mini-app-bot/internal/business/project"
)

type referenceAuth struct{}

func (referenceAuth) Authenticate(http.ResponseWriter, *http.Request) (string, string, bool) {
	return "owner", "", true
}

type referenceGroups struct{ feature.Groups }

func (referenceGroups) Get(_ context.Context, _, id string) (groupdomain.Group, error) {
	return groupdomain.Group{ID: id}, nil
}

type referenceSources struct{}

func (referenceSources) Authorize(_ context.Context, _, rid, sid string) (feature.Source, error) {
	return feature.Source{ID: rid, OwnerID: sid}, nil
}

type referenceMemory struct {
	feature.Store
	project feature.Project
	calls   int
}

func (m *referenceMemory) Get(context.Context, feature.Scope, string) (feature.Project, error) {
	return m.project, nil
}

func (m *referenceMemory) ChangeReference(_ context.Context, _ feature.Scope, _, _ string, ref feature.Reference, _ bool, _ int64) (feature.Project, error) {
	m.calls++
	if !feature.ValidReference(ref) {
		return feature.Project{}, feature.ErrInput
	}
	return m.project, nil
}

func (m *referenceMemory) Progress(context.Context, feature.Scope, *feature.Project) error {
	return nil
}

func (m *referenceMemory) ReferencePresent(context.Context, feature.Scope, string, string, feature.Reference) (bool, error) {
	return true, nil
}

func TestReferencePostHasIdentityRetryWithoutCreationHeader(t *testing.T) {
	gid, rid, pid := "C"+strings.Repeat("a", 32), strings.Repeat("b", 32), strings.Repeat("c", 32)
	m := &referenceMemory{project: feature.Project{ID: pid, DefaultRepositoryID: rid, OwnerID: gid, CreatorID: "owner", Title: "Project", Version: 1}}
	svc := &feature.Service{Groups: referenceGroups{}, Sources: referenceSources{}, Store: m}
	mux := http.NewServeMux()
	Register(mux, svc, referenceAuth{})
	base := "/api/v1/owners/" + gid + "/projects"
	body := `{"sourceType":"object","sourceId":"` + strings.Repeat("d", 32) + `","expectedVersion":1}`
	req := httptest.NewRequest("POST", base+"/"+pid+"/references", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	mux.ServeHTTP(w, req)
	if w.Code != 200 || m.calls != 1 {
		t.Fatal(w.Code, w.Body.String())
	}
	req = httptest.NewRequest("POST", base, strings.NewReader(`{"title":"new","body":"","defaultRepositoryId":"`+rid+`"}`))
	req.Header.Set("Content-Type", "application/json")
	w = httptest.NewRecorder()
	mux.ServeHTTP(w, req)
	if w.Code != 400 || m.calls != 1 {
		t.Fatal("project creation lost key requirement", w.Code)
	}
	req = httptest.NewRequest("GET", base+"/"+pid+"/reference-state?itemId="+strings.Repeat("d", 32)+"&desiredPresent=true", nil)
	w = httptest.NewRecorder()
	mux.ServeHTTP(w, req)
	if w.Code != 200 || !strings.Contains(w.Body.String(), `"matches":true`) || w.Header().Get("Cache-Control") != "no-store" {
		t.Fatal(w.Code, w.Body.String())
	}
}
