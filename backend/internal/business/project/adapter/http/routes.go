package projecthttp

import (
	projectApplication "github.com/963sup/line-mini-app-bot/internal/business/project/application"
	projectDomainEntity "github.com/963sup/line-mini-app-bot/internal/business/project/domain/entity"
)

import (
	"context"
	"errors"
	"mime"
	"net/http"
	"strings"
	"time"

	"github.com/963sup/line-mini-app-bot/internal/identity"
	transport "github.com/963sup/line-mini-app-bot/internal/transport/http"
	"github.com/963sup/line-mini-app-bot/internal/transport/http/api"
	"github.com/963sup/line-mini-app-bot/internal/transport/http/api/generated"
	"github.com/getkin/kin-openapi/openapi3filter"
	middleware "github.com/oapi-codegen/nethttp-middleware"
)

type (
	Authenticator interface {
		Authenticate(http.ResponseWriter, *http.Request) (string, string, bool)
	}
	projectServer struct{ service *projectApplication.Service }
)

var _ httpapi.ProjectServer = (*projectServer)(nil)

func Register(mux *http.ServeMux, service *projectApplication.Service, auth Authenticator) {
	spec, err := httpapi.Spec()
	if err != nil {
		panic(err)
	}
	wrap := func(next http.Handler) http.Handler {
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
			if !transport.ValidateUniqueQuery(w, r) {
				return
			}
			if r.Method == http.MethodPost || r.Method == http.MethodPatch {
				media, _, err := mime.ParseMediaType(r.Header.Get("Content-Type"))
				if err != nil || media != "application/json" {
					transport.WriteAPIError(w, 415, "invalid_content_type", "請使用 JSON。")
					return
				}
				r.Body = http.MaxBytesReader(w, r.Body, 65536)
				if r.Method == http.MethodPost && !strings.HasSuffix(r.URL.Path, "/items") && !strings.HasSuffix(r.URL.Path, "/references") && !strings.HasSuffix(r.URL.Path, "/inspection-submissions") && len(r.Header.Values("Idempotency-Key")) != 1 {
					transport.WriteAPIError(w, 400, "invalid_input", "請提供唯一的重試識別碼。")
					return
				}
			}
			ctx, cancel := context.WithTimeout(r.Context(), 12*time.Second)
			defer cancel()
			validate.ServeHTTP(w, r.WithContext(identity.WithUser(ctx, identity.User{ID: user, Name: name})))
		}))
	}
	httpapi.RegisterProjects(mux, &projectServer{service}, wrap)
	httpapi.RegisterOwnerProjects(mux, &projectServer{service}, wrap)
}

func view(i projectDomainEntity.Project) generated.Project {
	if i.LinkedRepositoryIDs == nil {
		i.LinkedRepositoryIDs = []string{}
	}
	return generated.Project{Id: i.ID, Title: i.Title, Body: i.Body, CreatorId: i.CreatorID, Version: i.Version, CreatedAt: i.CreatedAt, UpdatedAt: i.UpdatedAt, IssueCount: i.IssueCount, CompletedIssueCount: i.CompletedIssueCount, OwnerId: i.OwnerID, DefaultRepositoryId: i.DefaultRepositoryID, LinkedRepositoryIds: &i.LinkedRepositoryIDs, SourceAvailable: &i.SourceAvailable}
}

func writeError(w http.ResponseWriter, e error) bool {
	if e == nil {
		return false
	}
	code, status, message := "unavailable", 503, "結果尚未確認，請重新讀取；建立請沿用原鍵重試。"
	switch {
	case errors.Is(e, projectDomainEntity.ErrInput):
		code, status, message = "invalid_input", 400, "輸入或來源引用無效。"
	case errors.Is(e, projectDomainEntity.ErrNotFound):
		code, status, message = "not_found", 404, "找不到此資源。"
	case errors.Is(e, projectDomainEntity.ErrForbidden):
		code, status, message = "forbidden", 403, "只有建立者可以修改。"
	case errors.Is(e, projectDomainEntity.ErrInspectionPending):
		code, status, message = "conflict", 409, "已有待審送件，請查看驗收紀錄。"
	case errors.Is(e, projectDomainEntity.ErrInspectionHistory):
		code, status, message = "conflict", 409, "此引用有驗收紀錄，無法移除。"
	case errors.Is(e, projectDomainEntity.ErrConflict):
		code, status, message = "version_conflict", 409, "版本已更新，請重新讀取。"
	case errors.Is(e, projectDomainEntity.ErrCreationConflict):
		code, status, message = "idempotency_conflict", 409, "建立識別碼已用於不同內容，請確認原紀錄。"
	}
	transport.WriteServiceError(w, status, code, message, e)
	return true
}

func (s *projectServer) ListProjects(w http.ResponseWriter, r *http.Request, repositoryId string, p generated.ListProjectsParams) {
	cursor := ""
	if p.Cursor != nil {
		cursor = *p.Cursor
	}
	page, e := s.service.List(r.Context(), identity.UserID(r.Context()), repositoryId, cursor)
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
	i, e := s.service.Get(r.Context(), identity.UserID(r.Context()), repositoryId, id)
	if !writeError(w, e) {
		transport.WriteJSON(w, 200, view(i))
	}
}

func (s *projectServer) CreateProject(w http.ResponseWriter, r *http.Request, repositoryId string, p generated.CreateProjectParams) {
	var input generated.ProjectCreation
	if !transport.ReadJSON(w, r, &input, "資料無效。") {
		return
	}
	i, e := s.service.Create(r.Context(), identity.UserID(r.Context()), repositoryId, p.IdempotencyKey, projectDomainEntity.Content{Title: input.Title, Body: input.Body})
	if !writeError(w, e) {
		transport.WriteJSON(w, 200, view(i))
	}
}

func (s *projectServer) UpdateProject(w http.ResponseWriter, r *http.Request, repositoryId, id string, _ generated.UpdateProjectParams) {
	var input generated.ProjectUpdate
	if !transport.ReadJSON(w, r, &input, "資料無效。") {
		return
	}
	update := projectDomainEntity.Update{Content: projectDomainEntity.Content{Title: input.Title, Body: input.Body}, ExpectedVersion: input.ExpectedVersion}
	if input.ItemProgress != nil {
		update.ItemID, update.ItemStatus = input.ItemProgress.ItemId, string(input.ItemProgress.Status)
	}
	if input.LinkRepositoryId != nil {
		update.LinkRepositoryID = *input.LinkRepositoryId
	}
	if input.UnlinkRepositoryId != nil {
		update.UnlinkRepositoryID = *input.UnlinkRepositoryId
	}
	i, e := s.service.Update(r.Context(), identity.UserID(r.Context()), repositoryId, id, update)
	if !writeError(w, e) {
		transport.WriteJSON(w, 200, view(i))
	}
}

func (s *projectServer) ListProjectItems(w http.ResponseWriter, r *http.Request, repositoryId, id string, p generated.ListProjectItemsParams) {
	cursor := ""
	if p.Cursor != nil {
		cursor = *p.Cursor
	}
	page, e := s.service.ListItems(r.Context(), identity.UserID(r.Context()), repositoryId, id, cursor)
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
	if !transport.ReadJSON(w, r, &input, "資料無效。") {
		return
	}
	i, e := s.service.ChangeItem(r.Context(), identity.UserID(r.Context()), repositoryId, id, input.IssueId, true, input.ExpectedVersion)
	if !writeError(w, e) {
		transport.WriteJSON(w, 200, view(i))
	}
}

func (s *projectServer) RemoveProjectItem(w http.ResponseWriter, r *http.Request, repositoryId, id, itemId string, p generated.RemoveProjectItemParams) {
	i, e := s.service.ChangeItem(r.Context(), identity.UserID(r.Context()), repositoryId, id, itemId, false, p.ExpectedVersion)
	if !writeError(w, e) {
		transport.WriteJSON(w, 200, view(i))
	}
}

func (s *projectServer) ListOwnerProjects(w http.ResponseWriter, r *http.Request, ownerId string, p generated.ListOwnerProjectsParams) {
	s.ListProjects(w, r, ownerId, generated.ListProjectsParams(p))
}

func (s *projectServer) GetOwnerProject(w http.ResponseWriter, r *http.Request, ownerId, id string) {
	s.GetProject(w, r, ownerId, id)
}

func (s *projectServer) CreateOwnerProject(w http.ResponseWriter, r *http.Request, ownerId string, p generated.CreateOwnerProjectParams) {
	var input generated.OwnerProjectCreation
	if !transport.ReadJSON(w, r, &input, "資料無效。") {
		return
	}
	value, e := s.service.Create(r.Context(), identity.UserID(r.Context()), ownerId, p.IdempotencyKey, projectDomainEntity.Content{Title: input.Title, Body: input.Body, SourceRepositoryID: input.DefaultRepositoryId})
	if !writeError(w, e) {
		transport.WriteJSON(w, 200, view(value))
	}
}

func (s *projectServer) UpdateOwnerProject(w http.ResponseWriter, r *http.Request, ownerId, id string, _ generated.UpdateOwnerProjectParams) {
	s.UpdateProject(w, r, ownerId, id, generated.UpdateProjectParams{})
}

func (s *projectServer) ListOwnerProjectItems(w http.ResponseWriter, r *http.Request, ownerId, id string, p generated.ListOwnerProjectItemsParams) {
	s.ListProjectItems(w, r, ownerId, id, generated.ListProjectItemsParams(p))
}

func (s *projectServer) AddOwnerProjectItem(w http.ResponseWriter, r *http.Request, ownerId, id string, _ generated.AddOwnerProjectItemParams) {
	s.AddProjectItem(w, r, ownerId, id, generated.AddProjectItemParams{})
}

func (s *projectServer) RemoveOwnerProjectItem(w http.ResponseWriter, r *http.Request, ownerId, id, itemId string, p generated.RemoveOwnerProjectItemParams) {
	s.RemoveProjectItem(w, r, ownerId, id, itemId, generated.RemoveProjectItemParams{ExpectedVersion: p.ExpectedVersion})
}
