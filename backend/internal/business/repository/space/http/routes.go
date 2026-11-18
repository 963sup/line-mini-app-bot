package spacehttp

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"mime"
	"net/http"
	"time"

	feature "github.com/963sup/line-mini-app-bot/internal/business/repository/space"
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
	spaceServer struct{ service *feature.Service }
)

var _ httpapi.RepositorySpaceServer = (*spaceServer)(nil)

func Register(mux *http.ServeMux, service *feature.Service, auth Authenticator) {
	spec, err := httpapi.Spec()
	if err != nil {
		panic(err)
	}
	httpapi.RegisterRepositorySpaces(mux, &spaceServer{service}, func(next http.Handler) http.Handler {
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
				transport.WriteAPIError(w, 400, "invalid_input", "請求資料不符合業務空間契約。")
			},
		})(next)
		return http.NewCrossOriginProtection().Handler(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Cache-Control", "no-store")
			user, name, ok := auth.Authenticate(w, r)
			if !ok {
				return
			}
			if service == nil {
				transport.WriteAPIError(w, 503, "not_configured", "業務空間尚未啟用。")
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

func view(i feature.Space) generated.RepositorySpace {
	return generated.RepositorySpace{Id: i.ID, RepositoryId: i.RepositoryID, Name: i.Name, CreatorId: i.CreatorID, Version: i.Version, CreatedAt: i.CreatedAt, UpdatedAt: i.UpdatedAt}
}

func actor(r *http.Request) string { u, _ := identity.FromContext(r.Context()); return u.ID }
func writeError(w http.ResponseWriter, e error) bool {
	if e == nil {
		return false
	}
	code, httpStatus, message := "unavailable", 503, "業務空間結果尚未確認，請重新讀取；建立重試請沿用原識別碼。"
	switch {
	case errors.Is(e, feature.ErrNotFound):
		code, httpStatus, message = "not_found", 404, "找不到此空間或業務空間。"
	case errors.Is(e, feature.ErrInput):
		code, httpStatus, message = "invalid_input", 400, "請確認業務空間名稱。"
	case errors.Is(e, feature.ErrCreationConflict):
		code, httpStatus, message = "idempotency_conflict", 409, "重試識別碼已用於不同內容，請先確認原業務空間。"
	}
	transport.WriteAPIError(w, httpStatus, code, message)
	return true
}

func body(w http.ResponseWriter, r *http.Request, target any) bool {
	d := json.NewDecoder(r.Body)
	d.DisallowUnknownFields()
	if d.Decode(target) != nil || d.Decode(new(any)) != io.EOF {
		transport.WriteAPIError(w, 400, "invalid_input", "業務空間資料無效。")
		return false
	}
	return true
}

func (s *spaceServer) ListRepositorySpaces(w http.ResponseWriter, r *http.Request, rid string, p generated.ListRepositorySpacesParams) {
	cursor := ""
	if p.Cursor != nil {
		cursor = *p.Cursor
	}
	page, e := s.service.List(r.Context(), actor(r), rid, cursor)
	if writeError(w, e) {
		return
	}
	items := make([]generated.RepositorySpace, 0, len(page.Items))
	for _, i := range page.Items {
		items = append(items, view(i))
	}
	transport.WriteJSON(w, 200, generated.RepositorySpacePage{Items: items, NextCursor: page.NextCursor})
}

func (s *spaceServer) GetRepositorySpace(w http.ResponseWriter, r *http.Request, rid, id string) {
	i, e := s.service.Get(r.Context(), actor(r), rid, id)
	if !writeError(w, e) {
		transport.WriteJSON(w, 200, view(i))
	}
}

func (s *spaceServer) CreateRepositorySpace(w http.ResponseWriter, r *http.Request, rid string, p generated.CreateRepositorySpaceParams) {
	var in generated.RepositorySpaceCreation
	if !body(w, r, &in) {
		return
	}
	item, err := s.service.Create(r.Context(), actor(r), rid, p.IdempotencyKey, in.Name)
	if !writeError(w, err) {
		transport.WriteJSON(w, 200, view(item))
	}
}
