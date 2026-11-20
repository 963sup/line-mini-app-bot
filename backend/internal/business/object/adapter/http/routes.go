package objecthttp

import (
	objectApplication "github.com/963sup/line-mini-app-bot/internal/business/object/application"
	objectDomainEntity "github.com/963sup/line-mini-app-bot/internal/business/object/domain/entity"
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
	objectServer struct{ service *objectApplication.Service }
)

var _ httpapi.ObjectServer = (*objectServer)(nil)

func Register(mux *http.ServeMux, service *objectApplication.Service, auth Authenticator) {
	spec, err := httpapi.Spec()
	if err != nil {
		panic(err)
	}
	httpapi.RegisterObjects(mux, &objectServer{service}, func(next http.Handler) http.Handler {
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
				transport.WriteAPIError(w, 400, "invalid_input", "請求資料不符合物件契約。")
			},
		})(next)
		return http.NewCrossOriginProtection().Handler(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Cache-Control", "no-store")
			user, name, ok := auth.Authenticate(w, r)
			if !ok {
				return
			}
			if service == nil {
				transport.WriteAPIError(w, 503, "not_configured", "物件尚未啟用。")
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

func view(i objectDomainEntity.Object, actor string) generated.Object {
	var completed, spaceID *string
	if i.SpaceID != "" {
		spaceID = &i.SpaceID
	}
	if i.CompletedQuantity != "" {
		completed = &i.CompletedQuantity
	}
	return generated.Object{RequiresProgressReset: i.RequiresProgressReset(), CanEdit: i.CanEdit(actor), SpaceId: spaceID, CompletedQuantity: completed, Id: i.ID, RepositoryId: i.RepositoryID, Name: i.Name, Description: i.Description, Type: i.Type, Quantity: i.Quantity, Unit: i.Unit, Status: generated.ObjectStatus(i.Status), SourceDocument: i.SourceDocument, SourceRevision: i.SourceRevision, SourceItemNo: i.SourceItemNo, SourcePages: i.SourcePages, MaterialCode: i.MaterialCode, CreatorId: i.CreatorID, Version: i.Version, CreatedAt: i.CreatedAt, UpdatedAt: i.UpdatedAt}
}

func writeError(w http.ResponseWriter, e error) bool {
	if e == nil {
		return false
	}
	code, httpStatus, message := "unavailable", 503, "物件結果尚未確認，請重新讀取；建立重試請沿用原識別碼。"
	switch {
	case errors.Is(e, objectDomainEntity.ErrNotFound):
		code, httpStatus, message = "not_found", 404, "找不到此空間或物件。"
	case errors.Is(e, objectDomainEntity.ErrForbidden):
		code, httpStatus, message = "forbidden", 403, "只有建立者可以修改此物件。"
	case errors.Is(e, objectDomainEntity.ErrProgressRemoval):
		code, httpStatus, message = "progress_removal", 400, "已有完成數量的施作不能改為非施作，請先確認並調整完成數量。"
	case errors.Is(e, objectDomainEntity.ErrInput):
		code, httpStatus, message = "invalid_input", 400, "請確認名稱、類型、正數數量、單位與狀態。"
	case errors.Is(e, objectDomainEntity.ErrConflict):
		code, httpStatus, message = "version_conflict", 409, "物件已更新，請重新讀取。"
	case errors.Is(e, objectDomainEntity.ErrCreationConflict):
		code, httpStatus, message = "idempotency_conflict", 409, "重試識別碼已用於不同內容，請先確認原物件。"
	}
	transport.WriteServiceError(w, httpStatus, code, message, e)
	return true
}

func (s *objectServer) ListObjects(w http.ResponseWriter, r *http.Request, rid string, p generated.ListObjectsParams) {
	cursor := ""
	if p.Cursor != nil {
		cursor = *p.Cursor
	}
	page, e := s.service.List(r.Context(), identity.UserID(r.Context()), rid, cursor)
	if writeError(w, e) {
		return
	}
	items := make([]generated.Object, 0, len(page.Items))
	for _, i := range page.Items {
		items = append(items, view(i, identity.UserID(r.Context())))
	}
	transport.WriteJSON(w, 200, generated.ObjectPage{Items: items, NextCursor: page.NextCursor})
}

func (s *objectServer) GetObject(w http.ResponseWriter, r *http.Request, rid, id string) {
	i, e := s.service.Get(r.Context(), identity.UserID(r.Context()), rid, id)
	if !writeError(w, e) {
		transport.WriteJSON(w, 200, view(i, identity.UserID(r.Context())))
	}
}

func value(v *string) string {
	if v == nil {
		return ""
	}
	return *v
}

func (s *objectServer) CreateObject(w http.ResponseWriter, r *http.Request, rid string, p generated.CreateObjectParams) {
	var in generated.ObjectCreation
	if !transport.ReadJSON(w, r, &in, "物件資料無效。") {
		return
	}
	c := objectDomainEntity.Content{SpaceID: value(in.SpaceId), CompletedQuantity: value(in.CompletedQuantity), Name: in.Name, Description: in.Description, Type: in.Type, Quantity: in.Quantity, Unit: in.Unit, SourceDocument: value(in.SourceDocument), SourceRevision: value(in.SourceRevision), SourceItemNo: value(in.SourceItemNo), SourcePages: value(in.SourcePages), MaterialCode: value(in.MaterialCode)}
	i, e := s.service.Create(r.Context(), identity.UserID(r.Context()), rid, p.IdempotencyKey, c)
	if !writeError(w, e) {
		transport.WriteJSON(w, 200, view(i, identity.UserID(r.Context())))
	}
}

func (s *objectServer) UpdateObject(w http.ResponseWriter, r *http.Request, rid, id string, _ generated.UpdateObjectParams) {
	var in generated.ObjectUpdate
	if !transport.ReadJSON(w, r, &in, "物件資料無效。") {
		return
	}
	u := objectDomainEntity.Update{SpaceID: in.SpaceId, Content: objectDomainEntity.Content{CompletedQuantity: value(in.CompletedQuantity), Name: in.Name, Description: in.Description, Type: in.Type, Quantity: in.Quantity, Unit: in.Unit, SourceDocument: value(in.SourceDocument), SourceRevision: value(in.SourceRevision), SourceItemNo: value(in.SourceItemNo), SourcePages: value(in.SourcePages), MaterialCode: value(in.MaterialCode)}, Status: statusValue(in.Status), CompletedQuantity: in.CompletedQuantity, ExpectedVersion: in.ExpectedVersion, Sources: objectDomainEntity.SourceUpdate{SourceDocument: in.SourceDocument, SourceRevision: in.SourceRevision, SourceItemNo: in.SourceItemNo, SourcePages: in.SourcePages, MaterialCode: in.MaterialCode}}
	i, e := s.service.Update(r.Context(), identity.UserID(r.Context()), rid, id, u)
	if !writeError(w, e) {
		transport.WriteJSON(w, 200, view(i, identity.UserID(r.Context())))
	}
}

func statusValue(status *generated.ObjectUpdateStatus) string {
	if status == nil {
		return ""
	}
	return string(*status)
}
