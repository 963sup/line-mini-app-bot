package issuehttp

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"mime"
	"net/http"
	"time"

	"github.com/963sup/line-mini-app-bot/internal/business/issue"
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
	issueServer struct{ service *issue.Service }
)

var _ httpapi.IssueServer = (*issueServer)(nil)

func Register(mux *http.ServeMux, service *issue.Service, auth Authenticator) {
	spec, err := httpapi.Spec()
	if err != nil {
		panic(err)
	}
	httpapi.RegisterIssues(mux, &issueServer{service}, func(next http.Handler) http.Handler {
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
				transport.WriteAPIError(w, 400, "invalid_input", "請求資料不符合事項契約。")
			},
		})(next)
		return http.NewCrossOriginProtection().Handler(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Cache-Control", "no-store")
			user, name, ok := auth.Authenticate(w, r)
			if !ok {
				return
			}
			if service == nil {
				transport.WriteAPIError(w, 503, "not_configured", "事項尚未啟用。")
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
				if r.Method == http.MethodPost && len(r.Header.Values("Idempotency-Key")) != 1 {
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

func view(i issue.Issue) generated.Issue {
	ids := append([]string{}, i.LabelIDs...)
	var milestone *string
	if i.MilestoneID != "" {
		v := i.MilestoneID
		milestone = &v
	}
	return generated.Issue{MilestoneId: milestone, Id: i.ID, RepositoryId: i.RepositoryID, Number: i.Number, Title: i.Title, Body: i.Body, State: generated.IssueState(i.State), CreatorId: i.CreatorID, Version: i.Version, CreatedAt: i.CreatedAt, UpdatedAt: i.UpdatedAt, LabelIds: &ids}
}
func actor(r *http.Request) string { user, _ := identity.FromContext(r.Context()); return user.ID }
func writeError(w http.ResponseWriter, err error) bool {
	if err == nil {
		return false
	}
	code, status, message := "unavailable", 503, "事項結果尚未確認，請重新讀取；建立重試請沿用原識別碼。"
	switch {
	case errors.Is(err, issue.ErrNotFound):
		code, status, message = "not_found", 404, "找不到此空間或事項。"
	case errors.Is(err, issue.ErrForbidden):
		code, status, message = "forbidden", 403, "只有建立者可以修改此事項。"
	case errors.Is(err, issue.ErrInput):
		code, status, message = "invalid_input", 400, "請確認標題、內容、標籤與合法狀態轉換。"
	case errors.Is(err, issue.ErrConflict):
		code, status, message = "version_conflict", 409, "事項已更新，請重新讀取後再操作。"
	case errors.Is(err, issue.ErrCreationConflict):
		code, status, message = "idempotency_conflict", 409, "重試識別碼已用於不同內容，請先確認原事項。"
	}
	transport.WriteAPIError(w, status, code, message)
	return true
}

func body(w http.ResponseWriter, r *http.Request, target any) bool {
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()
	if decoder.Decode(target) != nil || decoder.Decode(new(any)) != io.EOF {
		transport.WriteAPIError(w, 400, "invalid_input", "事項資料無效。")
		return false
	}
	return true
}

func (s *issueServer) ListIssues(w http.ResponseWriter, r *http.Request, repositoryId string, params generated.ListIssuesParams) {
	f := issue.Filter{}
	if params.Cursor != nil {
		f.Cursor = *params.Cursor
	}
	if params.State != nil {
		f.State = string(*params.State)
	}
	if params.Query != nil {
		f.Query = *params.Query
	}
	page, err := s.service.List(r.Context(), actor(r), repositoryId, f)
	if writeError(w, err) {
		return
	}
	items := make([]generated.Issue, 0, len(page.Items))
	for _, i := range page.Items {
		items = append(items, view(i))
	}
	transport.WriteJSON(w, 200, generated.IssuePage{Items: items, NextCursor: page.NextCursor})
}

func (s *issueServer) GetIssue(w http.ResponseWriter, r *http.Request, repositoryId string, id string) {
	i, err := s.service.Get(r.Context(), actor(r), repositoryId, id)
	if !writeError(w, err) {
		transport.WriteJSON(w, 200, view(i))
	}
}

func (s *issueServer) CreateIssue(w http.ResponseWriter, r *http.Request, repositoryId string, params generated.CreateIssueParams) {
	var input generated.IssueCreation
	if !body(w, r, &input) {
		return
	}
	i, err := s.service.Create(r.Context(), actor(r), repositoryId, params.IdempotencyKey, issue.Content{Title: input.Title, Body: input.Body})
	if !writeError(w, err) {
		transport.WriteJSON(w, 200, view(i))
	}
}

func (s *issueServer) UpdateIssue(w http.ResponseWriter, r *http.Request, repositoryId string, id string, _ generated.UpdateIssueParams) {
	var input struct {
		generated.IssueUpdate
		MilestoneID json.RawMessage `json:"milestoneId"`
	}
	if !body(w, r, &input) {
		return
	}
	var milestone *string
	if len(input.MilestoneID) > 0 {
		v := ""
		if string(input.MilestoneID) != "null" && json.Unmarshal(input.MilestoneID, &v) != nil {
			transport.WriteAPIError(w, 400, "invalid_input", "里程碑資料無效。")
			return
		}
		milestone = &v
	}
	i, err := s.service.Update(r.Context(), actor(r), repositoryId, id, issue.Update{Content: issue.Content{Title: input.Title, Body: input.Body}, State: string(input.State), ExpectedVersion: input.ExpectedVersion, LabelIDs: input.LabelIds, MilestoneID: milestone})
	if !writeError(w, err) {
		transport.WriteJSON(w, 200, view(i))
	}
}
