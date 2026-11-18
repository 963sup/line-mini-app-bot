package projecthttp

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	feature "github.com/963sup/line-mini-app-bot/internal/business/project"
	"github.com/963sup/line-mini-app-bot/internal/business/space"
)

type referenceAuth struct{}

func (referenceAuth) Authenticate(http.ResponseWriter, *http.Request) (string, string, bool) {
	return "owner", "", true
}

type referenceSpaces struct{ feature.Spaces }

func (referenceSpaces) Get(_ context.Context, _, id string) (space.Space, error) {
	return space.Space{ID: id}, nil
}

type referenceSources struct{}

func (referenceSources) Authorize(_ context.Context, _, rid, sid string) (feature.Source, error) {
	return feature.Source{ID: rid, GroupID: sid}, nil
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

func TestReferencePostHasIdentityRetryWithoutCreationHeader(t *testing.T) {
	gid, rid, pid := "C"+strings.Repeat("a", 32), strings.Repeat("b", 32), strings.Repeat("c", 32)
	m := &referenceMemory{project: feature.Project{ID: pid, SourceRepositoryID: rid, SpaceID: gid, CreatorID: "owner", Title: "Project", Version: 1}}
	svc := &feature.Service{Spaces: referenceSpaces{}, Sources: referenceSources{}, Store: m}
	mux := http.NewServeMux()
	Register(mux, svc, referenceAuth{})
	base := "/api/v1/spaces/" + gid + "/projects"
	body := `{"sourceType":"object","sourceId":"` + strings.Repeat("d", 32) + `","expectedVersion":1}`
	req := httptest.NewRequest("POST", base+"/"+pid+"/references", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	mux.ServeHTTP(w, req)
	if w.Code != 200 || m.calls != 1 {
		t.Fatal(w.Code, w.Body.String())
	}
	req = httptest.NewRequest("POST", base, strings.NewReader(`{"title":"new","body":"","sourceRepositoryId":"`+rid+`"}`))
	req.Header.Set("Content-Type", "application/json")
	w = httptest.NewRecorder()
	mux.ServeHTTP(w, req)
	if w.Code != 400 || m.calls != 1 {
		t.Fatal("project creation lost key requirement", w.Code)
	}
}
