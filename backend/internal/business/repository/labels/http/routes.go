package labelhttp

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"mime"
	"net/http"
	"time"

	"github.com/963sup/line-mini-app-bot/internal/business/repository/labels"
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
	labelServer struct{ service *labels.Service }
)

var _ httpapi.LabelServer = (*labelServer)(nil)

func Register(mux *http.ServeMux, service *labels.Service, auth Authenticator) {
	spec, err := httpapi.Spec()
	if err != nil {
		panic(err)
	}
	httpapi.RegisterLabels(mux, &labelServer{service}, func(next http.Handler) http.Handler {
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
				transport.WriteAPIError(w, 400, "invalid_input", "請求資料不符合標籤契約。")
			},
		})(next)
		return http.NewCrossOriginProtection().Handler(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Cache-Control", "no-store")
			user, name, ok := auth.Authenticate(w, r)
			if !ok {
				return
			}
			if service == nil {
				transport.WriteAPIError(w, 503, "not_configured", "標籤尚未啟用。")
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
				r.Body = http.MaxBytesReader(w, r.Body, 1024)
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

func view(item labels.Label) generated.Label {
	return generated.Label{Id: item.ID, RepositoryId: item.RepositoryID, Name: item.Name, CreatedAt: item.CreatedAt}
}

func writeError(w http.ResponseWriter, err error) bool {
	if err == nil {
		return false
	}
	code, status, message := "unavailable", 503, "標籤結果尚未確認，請重新載入；建立重試請沿用原識別碼。"
	switch {
	case errors.Is(err, labels.ErrNotFound):
		code, status, message = "not_found", 404, "找不到此空間。"
	case errors.Is(err, labels.ErrInput):
		code, status, message = "invalid_input", 400, "請確認40字以內的名稱或有效分頁；名稱不得含控制字元。"
	case errors.Is(err, labels.ErrNameConflict):
		code, status, message = "name_conflict", 409, "此空間已有同名標籤，請重新載入標籤。"
	case errors.Is(err, labels.ErrCreationConflict):
		code, status, message = "idempotency_conflict", 409, "重試識別碼已用於不同名稱，請先確認原標籤。"
	}
	transport.WriteAPIError(w, status, code, message)
	return true
}

func (s *labelServer) ListLabels(w http.ResponseWriter, r *http.Request, repositoryId string, params generated.ListLabelsParams) {
	cursor := ""
	if params.Cursor != nil {
		cursor = *params.Cursor
	}
	page, err := s.service.List(r.Context(), actor(r), repositoryId, cursor)
	if writeError(w, err) {
		return
	}
	items := make([]generated.Label, 0, len(page.Items))
	for _, item := range page.Items {
		items = append(items, view(item))
	}
	transport.WriteJSON(w, 200, generated.LabelPage{Items: items, NextCursor: page.NextCursor})
}

func (s *labelServer) CreateLabel(w http.ResponseWriter, r *http.Request, repositoryId string, params generated.CreateLabelParams) {
	var input generated.LabelCreation
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()
	if decoder.Decode(&input) != nil || decoder.Decode(new(any)) != io.EOF {
		transport.WriteAPIError(w, 400, "invalid_input", "標籤資料無效。")
		return
	}
	item, err := s.service.Create(r.Context(), actor(r), repositoryId, params.IdempotencyKey, input.Name)
	if !writeError(w, err) {
		transport.WriteJSON(w, 200, view(item))
	}
}
func actor(r *http.Request) string { user, _ := identity.FromContext(r.Context()); return user.ID }
