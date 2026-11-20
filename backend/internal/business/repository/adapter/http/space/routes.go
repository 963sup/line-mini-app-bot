package spacehttp

import (
	repositoryApplicationSpace "github.com/963sup/line-mini-app-bot/internal/business/repository/application/space"
	repositoryDomainEntitySpace "github.com/963sup/line-mini-app-bot/internal/business/repository/domain/entity/space"
)

import (
	"context"
	"errors"
	"mime"
	"net/http"
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
	spaceServer struct {
		service *repositoryApplicationSpace.Service
	}
)

var _ httpapi.RepositorySpaceServer = (*spaceServer)(nil)

func Register(mux *http.ServeMux, service *repositoryApplicationSpace.Service, auth Authenticator) {
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
				transport.WriteAPIError(w, 400, "invalid_input", "請求資料不符合空間契約。")
			},
		})(next)
		return http.NewCrossOriginProtection().Handler(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Cache-Control", "no-store")
			user, name, ok := auth.Authenticate(w, r)
			if !ok {
				return
			}
			if service == nil {
				transport.WriteAPIError(w, 503, "not_configured", "空間尚未啟用。")
				return
			}
			if r.Method == http.MethodHead {
				r = r.Clone(r.Context())
				r.Method = http.MethodGet
			}
			if !transport.ValidateUniqueQuery(w, r) {
				return
			}
			if r.Method == http.MethodPost {
				media, _, err := mime.ParseMediaType(r.Header.Get("Content-Type"))
				if err != nil || media != "application/json" {
					transport.WriteAPIError(w, 415, "invalid_content_type", "請使用 JSON。")
					return
				}
				r.Body = http.MaxBytesReader(w, r.Body, 65536)
				if len(r.Header.Values("Idempotency-Key")) != 1 {
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

func view(i repositoryDomainEntitySpace.Space) generated.RepositorySpace {
	return generated.RepositorySpace{Id: i.ID, RepositoryId: i.RepositoryID, Name: i.Name, CreatorId: i.CreatorID, Version: i.Version, CreatedAt: i.CreatedAt, UpdatedAt: i.UpdatedAt}
}

func writeError(w http.ResponseWriter, e error) bool {
	if e == nil {
		return false
	}
	code, httpStatus, message := "unavailable", 503, "空間結果尚未確認，請重新讀取；建立重試請沿用原識別碼。"
	switch {
	case errors.Is(e, repositoryDomainEntitySpace.ErrNotFound):
		code, httpStatus, message = "not_found", 404, "找不到此儲存庫或空間。"
	case errors.Is(e, repositoryDomainEntitySpace.ErrInput):
		code, httpStatus, message = "invalid_input", 400, "請確認空間名稱。"
	case errors.Is(e, repositoryDomainEntitySpace.ErrCreationConflict):
		code, httpStatus, message = "idempotency_conflict", 409, "重試識別碼已用於不同內容，請先確認原空間。"
	}
	transport.WriteServiceError(w, httpStatus, code, message, e)
	return true
}

func (s *spaceServer) ListRepositorySpaces(w http.ResponseWriter, r *http.Request, rid string, p generated.ListRepositorySpacesParams) {
	cursor := ""
	if p.Cursor != nil {
		cursor = *p.Cursor
	}
	page, e := s.service.List(r.Context(), identity.UserID(r.Context()), rid, cursor)
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
	i, e := s.service.Get(r.Context(), identity.UserID(r.Context()), rid, id)
	if !writeError(w, e) {
		transport.WriteJSON(w, 200, view(i))
	}
}

func (s *spaceServer) CreateRepositorySpace(w http.ResponseWriter, r *http.Request, rid string, p generated.CreateRepositorySpaceParams) {
	var in generated.RepositorySpaceCreation
	if !transport.ReadJSON(w, r, &in, "空間資料無效。") {
		return
	}
	item, err := s.service.Create(r.Context(), identity.UserID(r.Context()), rid, p.IdempotencyKey, in.Name)
	if !writeError(w, err) {
		transport.WriteJSON(w, 200, view(item))
	}
}
