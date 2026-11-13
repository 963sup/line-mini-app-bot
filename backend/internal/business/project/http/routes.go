package projecthttp

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"mime"
	"net/http"
	"time"

	feature "github.com/963sup/line-mini-app-bot/internal/business/project"
	"github.com/963sup/line-mini-app-bot/internal/httpapi"
	"github.com/963sup/line-mini-app-bot/internal/httpapi/generated"
	"github.com/963sup/line-mini-app-bot/internal/identity"
	transport "github.com/963sup/line-mini-app-bot/internal/transport/http"
	"github.com/getkin/kin-openapi/openapi3filter"
	middleware "github.com/oapi-codegen/nethttp-middleware"
)

type (
	Authenticator interface {
		Authenticate(http.ResponseWriter, *http.Request) (string, string, bool)
	}
	projectServer struct{ service *feature.Service }
)

var _ httpapi.ProjectServer = (*projectServer)(nil)

func Register(mux *http.ServeMux, service *feature.Service, auth Authenticator) {
	spec, err := httpapi.Spec()
	if err != nil {
		panic(err)
	}
	httpapi.RegisterProjects(mux, &projectServer{service}, func(next http.Handler) http.Handler {
		document := *spec
		validate := middleware.OapiRequestValidatorWithOptions(&document, &middleware.Options{
			DoNotValidateServers: true,
			Options: openapi3filter.Options{AuthenticationFunc: func(ctx context.Context, _ *openapi3filter.AuthenticationInput) error {
				if _, ok := identity.FromContext(ctx); !ok {
					return errors.New("unverified identity")
				}
				return nil
			}},
			ErrorHandlerWithOpts: func(_ context.Context, _ error, w http.ResponseWriter, _ *http.Request, _ middleware.ErrorHandlerOpts) {
				transport.WriteAPIError(w, 400, "invalid_input", "請求資料不符合專案契約。")
			},
		})(next)
		return http.NewCrossOriginProtection().Handler(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Cache-Control", "no-store")
			user, name, ok := auth.Authenticate(w, r)
			if !ok {
				return
			}
			if service == nil {
				transport.WriteAPIError(w, 503, "not_configured", "專案尚未啟用。")
				return
			}
			if r.Method == http.MethodHead {
				r = r.Clone(r.Context())
				r.Method = http.MethodGet
			}
			for _, values := range r.URL.Query() {
				if len(values) > 1 {
					transport.WriteAPIError(w, 400, "invalid_input", "請勿重複查詢參數。")
					return
				}
			}
			if r.Method == http.MethodPost || r.Method == http.MethodPatch {
				media, _, err := mime.ParseMediaType(r.Header.Get("Content-Type"))
				if err != nil || media != "application/json" {
					transport.WriteAPIError(w, 415, "invalid_content_type", "請使用 JSON。")
					return
				}
				r.Body = http.MaxBytesReader(w, r.Body, 65536)
				if r.Method == http.MethodPost && r.URL.Path[len(r.URL.Path)-6:] != "/items" && len(r.Header.Values("Idempotency-Key")) != 1 {
					transport.WriteAPIError(w, 400, "invalid_input", "請提供唯一的重試識別碼。")
					return
				}
			}
			ctx, cancel := context.WithTimeout(r.Context(), 12*time.Second)
			defer cancel()
			validate.ServeHTTP(w, r.WithContext(identity.WithUser(ctx, identity.User{ID: user, Name: name})))
		}))
	})
}

func view(i feature.Project) generated.Project {
	return generated.Project{Id: i.ID, RepositoryId: i.RepositoryID, Title: i.Title, Body: i.Body, CreatorId: i.CreatorID, Version: i.Version, CreatedAt: i.CreatedAt, UpdatedAt: i.UpdatedAt, IssueCount: i.IssueCount, CompletedIssueCount: i.CompletedIssueCount}
}
func actor(r *http.Request) string { u, _ := identity.FromContext(r.Context()); return u.ID }
func body(w http.ResponseWriter, r *http.Request, target any) bool {
	d := json.NewDecoder(r.Body)
	d.DisallowUnknownFields()
	if d.Decode(target) != nil || d.Decode(new(any)) != io.EOF {
		transport.WriteAPIError(w, 400, "invalid_input", "資料無效。")
		return false
	}
	return true
}

func writeError(w http.ResponseWriter, e error) bool {
	if e == nil {
		return false
	}
	code, status, message := "unavailable", 503, "結果尚未確認，請重新讀取；建立請沿用原鍵重試。"
	switch {
	case errors.Is(e, feature.ErrInput):
		code, status, message = "invalid_input", 400, "輸入或同空間引用無效。"
	case errors.Is(e, feature.ErrNotFound):
		code, status, message = "not_found", 404, "找不到此資源。"
	case errors.Is(e, feature.ErrForbidden):
		code, status, message = "forbidden", 403, "只有建立者可以修改。"
	case errors.Is(e, feature.ErrConflict):
		code, status, message = "version_conflict", 409, "版本已更新，請重新讀取。"
	case errors.Is(e, feature.ErrCreationConflict):
		code, status, message = "idempotency_conflict", 409, "建立識別碼已用於不同內容，請確認原紀錄。"
	}
	transport.WriteAPIError(w, status, code, message)
	return true
}

func (s *projectServer) ListProjects(w http.ResponseWriter, r *http.Request, repositoryId string, p generated.ListProjectsParams) {
	cursor := ""
	if p.Cursor != nil {
		cursor = *p.Cursor
	}
	page, e := s.service.List(r.Context(), actor(r), repositoryId, cursor)
	if writeError(w, e) {
		return
	}
	items := make([]generated.Project, 0, len(page.Items))
	for _, i := range page.Items {
		items = append(items, view(i))
	}
	transport.WriteJSON(w, 200, generated.ProjectPage{Items: items, NextCursor: page.NextCursor})
}

func (s *projectServer) GetProject(w http.ResponseWriter, r *http.Request, repositoryId, id string) {
	i, e := s.service.Get(r.Context(), actor(r), repositoryId, id)
	if !writeError(w, e) {
		transport.WriteJSON(w, 200, view(i))
	}
}

func (s *projectServer) CreateProject(w http.ResponseWriter, r *http.Request, repositoryId string, p generated.CreateProjectParams) {
	var input generated.ProjectCreation
	if !body(w, r, &input) {
		return
	}
	i, e := s.service.Create(r.Context(), actor(r), repositoryId, p.IdempotencyKey, feature.Content{Title: input.Title, Body: input.Body})
	if !writeError(w, e) {
		transport.WriteJSON(w, 200, view(i))
	}
}

func (s *projectServer) UpdateProject(w http.ResponseWriter, r *http.Request, repositoryId, id string, _ generated.UpdateProjectParams) {
	var input generated.ProjectUpdate
	if !body(w, r, &input) {
		return
	}
	i, e := s.service.Update(r.Context(), actor(r), repositoryId, id, feature.Update{Content: feature.Content{Title: input.Title, Body: input.Body}, ExpectedVersion: input.ExpectedVersion})
	if !writeError(w, e) {
		transport.WriteJSON(w, 200, view(i))
	}
}

func (s *projectServer) ListProjectItems(w http.ResponseWriter, r *http.Request, repositoryId, id string, p generated.ListProjectItemsParams) {
	cursor := ""
	if p.Cursor != nil {
		cursor = *p.Cursor
	}
	page, e := s.service.ListItems(r.Context(), actor(r), repositoryId, id, cursor)
	if writeError(w, e) {
		return
	}
	items := make([]generated.ProjectItem, 0, len(page.Items))
	for _, i := range page.Items {
		items = append(items, generated.ProjectItem{Id: i.ID, IssueId: i.IssueID, Title: i.Title, State: generated.ProjectItemState(i.State)})
	}
	transport.WriteJSON(w, 200, generated.ProjectItemPage{Items: items, NextCursor: page.NextCursor})
}

func (s *projectServer) AddProjectItem(w http.ResponseWriter, r *http.Request, repositoryId, id string, _ generated.AddProjectItemParams) {
	var input generated.ProjectItemAddition
	if !body(w, r, &input) {
		return
	}
	i, e := s.service.ChangeItem(r.Context(), actor(r), repositoryId, id, input.IssueId, true, input.ExpectedVersion)
	if !writeError(w, e) {
		transport.WriteJSON(w, 200, view(i))
	}
}

func (s *projectServer) RemoveProjectItem(w http.ResponseWriter, r *http.Request, repositoryId, id, itemId string, p generated.RemoveProjectItemParams) {
	i, e := s.service.ChangeItem(r.Context(), actor(r), repositoryId, id, itemId, false, p.ExpectedVersion)
	if !writeError(w, e) {
		transport.WriteJSON(w, 200, view(i))
	}
}
