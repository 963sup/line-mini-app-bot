package projecthttp

import (
	projectApplication "github.com/963sup/line-mini-app-bot/internal/business/project/application"
	projectApplicationPort "github.com/963sup/line-mini-app-bot/internal/business/project/application/port"
	projectDomainEntity "github.com/963sup/line-mini-app-bot/internal/business/project/domain/entity"
)

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

type referenceAuth struct{}

func (referenceAuth) Authenticate(http.ResponseWriter, *http.Request) (string, string, bool) {
	return "owner", "", true
}

type referenceGroups struct{}

func (referenceGroups) AuthorizeGroup(_ context.Context, _, id string) (string, error) {
	return id, nil
}

type referenceSources struct{}

func (referenceSources) Authorize(_ context.Context, _, rid, sid string) (projectDomainEntity.Source, error) {
	return projectDomainEntity.Source{ID: rid, OwnerID: sid}, nil
}

type referenceMemory struct {
	projectApplicationPort.Store
	project projectDomainEntity.Project
	calls   int
}

func (m *referenceMemory) Get(context.Context, projectApplicationPort.Scope, string) (projectDomainEntity.Project, error) {
	return m.project, nil
}

func (m *referenceMemory) ChangeReference(_ context.Context, _ projectApplicationPort.Scope, _, _ string, ref projectDomainEntity.Reference, _ bool, _ int64) (projectDomainEntity.Project, error) {
	m.calls++
	if !projectDomainEntity.ValidReference(ref) {
		return projectDomainEntity.Project{}, projectDomainEntity.ErrInput
	}
	return m.project, nil
}

func (m *referenceMemory) Progress(context.Context, projectApplicationPort.Scope, *projectDomainEntity.Project) error {
	return nil
}

func (m *referenceMemory) ReferencePresent(context.Context, projectApplicationPort.Scope, string, string, projectDomainEntity.Reference) (bool, error) {
	return true, nil
}

func TestReferencePostHasIdentityRetryWithoutCreationHeader(t *testing.T) {
	gid, rid, pid := "aaaaaaaa-aaaa-4aaa-8aaa-aaaaaaaaaaaa", strings.Repeat("b", 32), strings.Repeat("c", 32)
	m := &referenceMemory{project: projectDomainEntity.Project{ID: pid, DefaultRepositoryID: rid, OwnerID: gid, CreatorID: "owner", Title: "Project", Version: 1}}
	svc := &projectApplication.Service{AuthorizeOrganization: func(context.Context, string, string) error { return nil }, Groups: referenceGroups{}, Sources: referenceSources{}, Store: m}
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

func TestReferenceInspectionModes(t *testing.T) {
	for _, tc := range []struct {
		kind        string
		unavailable bool
		want        string
	}{
		{"object", false, "submit_if_empty"}, {"object", true, "history_only"}, {"discussion", false, "none"}, {"issue", false, "none"},
	} {
		ref := projectDomainEntity.Reference{SourceType: tc.kind, SourceUnavailable: tc.unavailable}
		if got := referenceView(ref).InspectionMode; string(got) != tc.want {
			t.Errorf("%s unavailable=%t: %s", tc.kind, tc.unavailable, got)
		}
	}
}
