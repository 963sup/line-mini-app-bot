package issuehttp

import (
	"context"
	"encoding/json"
	"errors"
	"mime"
	"net/http"
	"time"

	issueapplication "github.com/963sup/line-mini-app-bot/internal/business/issue/application"
	issuedto "github.com/963sup/line-mini-app-bot/internal/business/issue/application/dto"
	issueentity "github.com/963sup/line-mini-app-bot/internal/business/issue/domain/entity"
	"github.com/963sup/line-mini-app-bot/internal/platform/identity"
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
	issueServer struct{ service *issueapplication.Service }
)

var _ httpapi.IssueServer = (*issueServer)(nil)

func Register(mux *http.ServeMux, service *issueapplication.Service, auth Authenticator) {
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
				transport.WriteRequestError(w, 400, "invalid_input", "請求資料不符合問題契約。", false)
			},
		})(next)
		return http.NewCrossOriginProtection().Handler(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Cache-Control", "no-store")
			user, name, ok := auth.Authenticate(w, r)
			if !ok {
				return
			}
			if service == nil {
				transport.WriteAPIError(w, 503, "not_configured", "問題尚未啟用。")
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
				if r.Method == http.MethodPost && len(r.Header.Values("Idempotency-Key")) != 1 {
					transport.WriteRequestError(w, 400, "invalid_input", "請提供唯一的重試識別碼。", false)
					return
				}
			}
			ctx, cancel := context.WithTimeout(r.Context(), 12*time.Second)
			defer cancel()
			validate.ServeHTTP(w, r.WithContext(identity.WithUser(ctx, identity.User{ID: user, Name: name})))
		}))
	})
}

func view(i issueentity.Issue, actor string) generated.Issue {
	ids := append([]string{}, i.LabelIDs...)
	var milestone *string
	if i.MilestoneID != "" {
		v := i.MilestoneID
		milestone = &v
	}
	return generated.Issue{CanEdit: i.CanEdit(actor), MilestoneId: milestone, Id: i.ID, RepositoryId: i.RepositoryID, Number: i.Number, Title: i.Title, Body: i.Body, State: generated.IssueState(i.State), CreatorId: i.CreatorID, Version: i.Version, CreatedAt: i.CreatedAt, UpdatedAt: i.UpdatedAt, LabelIds: &ids}
}

func writeError(w http.ResponseWriter, err error) bool {
	if err == nil {
		return false
	}
	code, status, message := "unavailable", 503, "問題結果尚未確認，請重新讀取；建立重試請沿用原識別碼。"
	switch {
	case errors.Is(err, issueentity.ErrNotFound):
		code, status, message = "not_found", 404, "找不到此空間或問題。"
	case errors.Is(err, issueentity.ErrForbidden):
		code, status, message = "forbidden", 403, "只有建立者可以修改此問題。"
	case errors.Is(err, issueentity.ErrInput):
		code, status, message = "invalid_input", 400, "請確認標題、內容、標籤與合法狀態轉換。"
	case errors.Is(err, issueentity.ErrConflict):
		code, status, message = "version_conflict", 409, "問題已更新，請重新讀取後再操作。"
	case errors.Is(err, issueentity.ErrCreationConflict):
		code, status, message = "idempotency_conflict", 409, "重試識別碼已用於不同內容，請先確認原問題。"
	}
	transport.WriteResourceError(w, status, code, message, err, code == "unavailable")
	return true
}

func (s *issueServer) ListIssues(w http.ResponseWriter, r *http.Request, repositoryId string, params generated.ListIssuesParams) {
	f := issuedto.Filter{}
	if params.Cursor != nil {
		f.Cursor = *params.Cursor
	}
	if params.State != nil {
		f.State = string(*params.State)
	}
	if params.Query != nil {
		f.Query = *params.Query
	}
	page, err := s.service.List(r.Context(), identity.UserID(r.Context()), repositoryId, f)
	if writeError(w, err) {
		return
	}
	items := make([]generated.Issue, 0, len(page.Items))
	for _, i := range page.Items {
		items = append(items, view(i, identity.UserID(r.Context())))
	}
	transport.WriteJSON(w, 200, generated.IssuePage{Items: items, NextCursor: page.NextCursor})
}

func (s *issueServer) GetIssue(w http.ResponseWriter, r *http.Request, repositoryId string, id string) {
	i, err := s.service.Get(r.Context(), identity.UserID(r.Context()), repositoryId, id)
	if !writeError(w, err) {
		transport.WriteJSON(w, 200, view(i, identity.UserID(r.Context())))
	}
}

func (s *issueServer) CreateIssue(w http.ResponseWriter, r *http.Request, repositoryId string, params generated.CreateIssueParams) {
	var input generated.IssueCreation
	if !transport.ReadJSON(w, r, &input, "問題資料無效。") {
		return
	}
	i, err := s.service.Create(r.Context(), identity.UserID(r.Context()), repositoryId, params.IdempotencyKey, issueentity.Content{Title: input.Title, Body: input.Body})
	if !writeError(w, err) {
		transport.WriteJSON(w, 200, view(i, identity.UserID(r.Context())))
	}
}

func (s *issueServer) UpdateIssue(w http.ResponseWriter, r *http.Request, repositoryId string, id string, _ generated.UpdateIssueParams) {
	var input struct {
		generated.IssueUpdate
		MilestoneID json.RawMessage `json:"milestoneId"`
	}
	if !transport.ReadJSON(w, r, &input, "問題資料無效。") {
		return
	}
	var milestone *string
	if len(input.MilestoneID) > 0 {
		v := ""
		if string(input.MilestoneID) != "null" && json.Unmarshal(input.MilestoneID, &v) != nil {
			transport.WriteRequestError(w, 400, "invalid_input", "里程碑資料無效。", false)
			return
		}
		milestone = &v
	}
	i, err := s.service.Update(r.Context(), identity.UserID(r.Context()), repositoryId, id, issueentity.Update{Content: issueentity.Content{Title: input.Title, Body: input.Body}, State: string(input.State), ExpectedVersion: input.ExpectedVersion, LabelIDs: input.LabelIds, MilestoneID: milestone})
	if !writeError(w, err) {
		transport.WriteJSON(w, 200, view(i, identity.UserID(r.Context())))
	}
}
