package subobjecthttp

import (
	subobjectApplication "github.com/963sup/line-mini-app-bot/internal/business/subobject/application"
	subobjectDomainEntity "github.com/963sup/line-mini-app-bot/internal/business/subobject/domain/entity"
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
	subobjectServer struct{ service *subobjectApplication.Service }
)

var _ httpapi.SubobjectServer = (*subobjectServer)(nil)

func Register(mux *http.ServeMux, service *subobjectApplication.Service, auth Authenticator) {
	spec, err := httpapi.Spec()
	if err != nil {
		panic(err)
	}
	httpapi.RegisterSubobjects(mux, &subobjectServer{service}, func(next http.Handler) http.Handler {
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
				transport.WriteAPIError(w, 400, "invalid_input", "請求資料不符合子物件契約。")
			},
		})(next)
		return http.NewCrossOriginProtection().Handler(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Cache-Control", "no-store")
			user, name, ok := auth.Authenticate(w, r)
			if !ok {
				return
			}
			if service == nil {
				transport.WriteAPIError(w, 503, "not_configured", "子物件尚未啟用。")
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

func view(i subobjectDomainEntity.Subobject) generated.Subobject {
	var completed *string
	if i.CompletedQuantity != "" {
		completed = &i.CompletedQuantity
	}
	return generated.Subobject{Id: i.ID, ParentObjectId: i.ParentObjectID, RepositoryId: i.RepositoryID, CreatorId: i.CreatorID, Name: i.Name, Type: generated.SubobjectType(i.Type), Quantity: i.Quantity, Unit: i.Unit, CompletedQuantity: completed, Status: generated.SubobjectStatus(i.Status), Version: i.Version, CreatedAt: i.CreatedAt, UpdatedAt: i.UpdatedAt}
}

func writeError(w http.ResponseWriter, e error) bool {
	if e == nil {
		return false
	}
	code, httpStatus, message := "unavailable", 503, "子物件結果尚未確認，請重新讀取；建立重試請沿用原識別碼。"
	switch {
	case errors.Is(e, subobjectDomainEntity.ErrNotFound):
		code, httpStatus, message = "not_found", 404, "找不到此空間或子物件。"
	case errors.Is(e, subobjectDomainEntity.ErrForbidden):
		code, httpStatus, message = "forbidden", 403, "只有父物件建立者可以管理子物件。"
	case errors.Is(e, subobjectDomainEntity.ErrProgressRemoval):
		code, httpStatus, message = "progress_removal", 400, "已有完成數量的施作不能改為非施作，請先確認並調整完成數量。"
	case errors.Is(e, subobjectDomainEntity.ErrInput):
		code, httpStatus, message = "invalid_input", 400, "請確認名稱、類型、總數量、單位與完成數量。"
	case errors.Is(e, subobjectDomainEntity.ErrConflict):
		code, httpStatus, message = "version_conflict", 409, "子物件已更新，請重新讀取。"
	case errors.Is(e, subobjectDomainEntity.ErrCreationConflict):
		code, httpStatus, message = "idempotency_conflict", 409, "重試識別碼已用於不同內容，請先確認原子物件。"
	}
	transport.WriteServiceError(w, httpStatus, code, message, e)
	return true
}

func (s *subobjectServer) ListSubobjects(w http.ResponseWriter, r *http.Request, repositoryID, parentObjectID string, p generated.ListSubobjectsParams) {
	cursor := ""
	if p.Cursor != nil {
		cursor = *p.Cursor
	}
	page, e := s.service.List(r.Context(), identity.UserID(r.Context()), repositoryID, parentObjectID, cursor)
	if writeError(w, e) {
		return
	}
	items := make([]generated.Subobject, 0, len(page.Items))
	for _, i := range page.Items {
		items = append(items, view(i))
	}
	transport.WriteJSON(w, 200, generated.SubobjectPage{Items: items, NextCursor: page.NextCursor})
}

func (s *subobjectServer) GetSubobject(w http.ResponseWriter, r *http.Request, repositoryID, parentObjectID, subobjectID string) {
	i, e := s.service.Get(r.Context(), identity.UserID(r.Context()), repositoryID, parentObjectID, subobjectID)
	if !writeError(w, e) {
		transport.WriteJSON(w, 200, view(i))
	}
}

func value(v *string) string {
	if v == nil {
		return ""
	}
	return *v
}

func (s *subobjectServer) CreateSubobject(w http.ResponseWriter, r *http.Request, repositoryID, parentObjectID string, p generated.CreateSubobjectParams) {
	var in generated.SubobjectCreation
	if !transport.ReadJSON(w, r, &in, "子物件資料無效。") {
		return
	}
	c := subobjectDomainEntity.Content{Name: in.Name, Type: string(in.Type), Quantity: in.Quantity, Unit: in.Unit, CompletedQuantity: value(in.CompletedQuantity)}
	i, e := s.service.Create(r.Context(), identity.UserID(r.Context()), repositoryID, parentObjectID, p.IdempotencyKey, c)
	if !writeError(w, e) {
		transport.WriteJSON(w, 200, view(i))
	}
}

func (s *subobjectServer) UpdateSubobject(w http.ResponseWriter, r *http.Request, repositoryID, parentObjectID, subobjectID string, _ generated.UpdateSubobjectParams) {
	var in generated.SubobjectUpdate
	if !transport.ReadJSON(w, r, &in, "子物件資料無效。") {
		return
	}
	u := subobjectDomainEntity.Update{Content: subobjectDomainEntity.Content{Name: in.Name, Type: string(in.Type), Quantity: in.Quantity, Unit: in.Unit}, CompletedQuantity: in.CompletedQuantity, ExpectedVersion: in.ExpectedVersion}
	i, e := s.service.Update(r.Context(), identity.UserID(r.Context()), repositoryID, parentObjectID, subobjectID, u)
	if !writeError(w, e) {
		transport.WriteJSON(w, 200, view(i))
	}
}
