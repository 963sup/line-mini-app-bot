package objecthttp

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/963sup/line-mini-app-bot/internal/business/object/application"
	"github.com/963sup/line-mini-app-bot/internal/business/object/application/dto"
	"github.com/963sup/line-mini-app-bot/internal/business/object/domain/entity"
	"github.com/963sup/line-mini-app-bot/internal/business/object/domain/template"
	repository "github.com/963sup/line-mini-app-bot/internal/business/repository/domain/entity"
	"github.com/963sup/line-mini-app-bot/internal/transport/http/api/generated"
)

type templateStore struct {
	snapshot dto.Template
	calls    int
}

func (s *templateStore) CreateTemplate(_ context.Context, r repository.Repository, actor, _ string, in dto.TemplateCreation) (dto.Template, error) {
	s.calls++
	s.snapshot = dto.Template{ID: strings.Repeat("b", 32), RepositoryID: r.ID, Name: in.Name, CreatorID: actor, Version: 1, Nodes: in.Nodes, CreatedAt: time.Now().UTC()}
	return s.snapshot, nil
}

func (s *templateStore) ListTemplates(context.Context, repository.Repository, string, string) (dto.TemplatePage, error) {
	s.calls++
	return dto.TemplatePage{Items: []dto.Template{s.snapshot}}, nil
}

func (s *templateStore) GetTemplate(context.Context, repository.Repository, string) (dto.Template, error) {
	s.calls++
	return s.snapshot, nil
}

func (s *templateStore) InstantiateTemplate(_ context.Context, _ repository.Repository, _, _ string, snapshot dto.Template, _ string, _ []entity.Content, _ ...[]template.Node) (dto.TemplateInstance, error) {
	s.calls++
	id := strings.Repeat("c", 32)
	return dto.TemplateInstance{RootObjectID: id, ObjectIDs: []string{id}, TemplateID: snapshot.ID, TemplateVersion: 1, Nodes: []dto.TemplateInstanceNode{{ObjectID: id, Name: snapshot.Nodes[0].Name}}}, nil
}

func TestTemplateHTTPContractAndAuthorization(t *testing.T) {
	rid := strings.Repeat("a", 32)
	tid := strings.Repeat("b", 32)
	templates := &templateStore{}
	mux := http.NewServeMux()
	service := &application.Service{Repositories: repositories{}, Store: &store{}, Templates: templates}
	Register(mux, service, auth{})
	invoke := func(method, path, body, key string) *httptest.ResponseRecorder {
		request := httptest.NewRequest(method, "/api/v1/repositories/"+rid+path, strings.NewReader(body))
		request.Header.Set("Content-Type", "application/json")
		if key != "" {
			request.Header.Set("Idempotency-Key", key)
		}
		response := httptest.NewRecorder()
		mux.ServeHTTP(response, request)
		return response
	}
	response := invoke("POST", "/object-templates", `{"name":"範本","nodes":[{"id":"root","name":"工程"}]}`, "create-template-key")
	if response.Code != 200 {
		t.Fatal(response.Code, response.Body.String())
	}
	var created generated.ObjectTemplate
	if e := json.Unmarshal(response.Body.Bytes(), &created); e != nil || created.Version != 1 || created.CreatorId != "A" || len(created.Nodes) != 1 {
		t.Fatal(created, e)
	}
	for _, path := range []string{"/object-templates", "/object-templates/" + tid} {
		if rr := invoke("GET", path, "", ""); rr.Code != 200 || rr.Header().Get("Cache-Control") != "no-store" {
			t.Fatal(rr.Code, rr.Body.String())
		}
	}
	rr := invoke("POST", "/object-templates/"+tid+"/instances", `{}`, "apply-template-key")
	var result generated.ObjectTemplateInstance
	if e := json.Unmarshal(rr.Body.Bytes(), &result); rr.Code != 200 || e != nil || len(result.Nodes) != 1 || result.RootObjectId != result.Nodes[0].ObjectId {
		t.Fatal(rr.Code, result, e)
	}
	before := templates.calls
	for _, body := range []string{`{"name":"x","nodes":[{"id":"root","name":"x"}],"creatorId":"forged"}`, `{"name":"x","nodes":[{"id":"r","name":"x"},{"id":"s","name":"y"}]}`} {
		rr := invoke("POST", "/object-templates", body, "invalid-template-key")
		if rr.Code != 400 {
			t.Fatal(rr.Code, rr.Body.String())
		}
	}
	if rr := invoke("POST", "/object-templates", `{"name":"x","nodes":[{"id":"r","name":"x"}]}`, ""); rr.Code != 400 {
		t.Fatal(rr.Code)
	}
	service.Repositories = repositories{deny: true}
	if rr := invoke("POST", "/object-templates/"+tid+"/instances", `{}`, "denied-template-key"); rr.Code != 404 {
		t.Fatal(rr.Code, rr.Body.String())
	}
	if templates.calls != before {
		t.Fatal("rejected input or denied authority reached store")
	}
}

type personalTemplateStore struct{ templateStore }

func (s *personalTemplateStore) CreatePersonalTemplate(ctx context.Context, actor, key string, in dto.TemplateCreation) (dto.Template, error) {
	return s.CreateTemplate(ctx, repository.Repository{}, actor, key, in)
}

func (s *personalTemplateStore) ListPersonalTemplates(ctx context.Context, actor, cursor string) (dto.TemplatePage, error) {
	return s.ListTemplates(ctx, repository.Repository{}, actor, cursor)
}

func (s *personalTemplateStore) InstantiatePersonalTemplate(context.Context, repository.Repository, string, string, string, string, []template.Node) (dto.TemplateInstance, error) {
	panic("not used by library operations")
}

func TestPersonalTemplateLibraryDoesNotRequireRepository(t *testing.T) {
	templates := &personalTemplateStore{}
	service := &application.Service{PersonalTemplates: templates}
	mux := http.NewServeMux()
	Register(mux, service, auth{})
	for _, method := range []string{"POST", "GET"} {
		request := httptest.NewRequest(method, "/api/v1/object-templates", strings.NewReader(`{"name":"範本","nodes":[{"id":"root","name":"工程"}]}`))
		request.Header.Set("Content-Type", "application/json")
		request.Header.Set("Idempotency-Key", "personal-template-key")
		response := httptest.NewRecorder()
		mux.ServeHTTP(response, request)
		if response.Code != http.StatusOK {
			t.Fatalf("%s: %d %s", method, response.Code, response.Body.String())
		}
	}
	if templates.calls != 2 || templates.snapshot.CreatorID != "A" {
		t.Fatalf("unexpected template calls: %+v", templates)
	}
	denied := http.NewServeMux()
	Register(denied, service, auth{deny: true})
	response := httptest.NewRecorder()
	denied.ServeHTTP(response, httptest.NewRequest("GET", "/api/v1/object-templates", nil))
	if response.Code != 401 || templates.calls != 2 {
		t.Fatal("unauthenticated request reached store")
	}
}
