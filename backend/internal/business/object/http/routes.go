package objecthttp

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"mime"
	"net/http"
	"time"

	"github.com/963sup/line-mini-app-bot/internal/business/object"
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
	objectServer struct{ service *object.Service }
)

var _ httpapi.ObjectServer = (*objectServer)(nil)

func Register(mux *http.ServeMux, service *object.Service, auth Authenticator) {
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
				transport.WriteAPIError(w, 400, "invalid_input", "請求資料不符合工程項目契約。")
			},
		})(next)
		return http.NewCrossOriginProtection().Handler(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Cache-Control", "no-store")
			user, name, ok := auth.Authenticate(w, r)
			if !ok {
				return
			}
			if service == nil {
				transport.WriteAPIError(w, 503, "not_configured", "工程項目尚未啟用。")
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

func view(i object.Object) generated.Object {
	return generated.Object{Id: i.ID, RepositoryId: i.RepositoryID, Name: i.Name, Description: i.Description, Type: i.Type, Quantity: i.Quantity, Unit: i.Unit, Status: generated.ObjectStatus(i.Status), SourceDocument: i.SourceDocument, SourceRevision: i.SourceRevision, SourceItemNo: i.SourceItemNo, SourcePages: i.SourcePages, MaterialCode: i.MaterialCode, CreatorId: i.CreatorID, Version: i.Version, CreatedAt: i.CreatedAt, UpdatedAt: i.UpdatedAt}
}
func actor(r *http.Request) string { u, _ := identity.FromContext(r.Context()); return u.ID }
func writeError(w http.ResponseWriter, e error) bool {
	if e == nil {
		return false
	}
	code, httpStatus, message := "unavailable", 503, "工程項目結果尚未確認，請重新讀取；建立重試請沿用原識別碼。"
	switch {
	case errors.Is(e, object.ErrNotFound):
		code, httpStatus, message = "not_found", 404, "找不到此空間或工程項目。"
	case errors.Is(e, object.ErrForbidden):
		code, httpStatus, message = "forbidden", 403, "只有建立者可以修改此工程項目。"
	case errors.Is(e, object.ErrInput):
		code, httpStatus, message = "invalid_input", 400, "請確認名稱、類型、正數數量、單位與狀態。"
	case errors.Is(e, object.ErrConflict):
		code, httpStatus, message = "version_conflict", 409, "工程項目已更新，請重新讀取。"
	case errors.Is(e, object.ErrCreationConflict):
		code, httpStatus, message = "idempotency_conflict", 409, "重試識別碼已用於不同內容，請先確認原工程項目。"
	}
	transport.WriteAPIError(w, httpStatus, code, message)
	return true
}

func body(w http.ResponseWriter, r *http.Request, target any) bool {
	d := json.NewDecoder(r.Body)
	d.DisallowUnknownFields()
	if d.Decode(target) != nil || d.Decode(new(any)) != io.EOF {
		transport.WriteAPIError(w, 400, "invalid_input", "工程項目資料無效。")
		return false
	}
	return true
}

func (s *objectServer) ListObjects(w http.ResponseWriter, r *http.Request, rid string, p generated.ListObjectsParams) {
	cursor := ""
	if p.Cursor != nil {
		cursor = *p.Cursor
	}
	page, e := s.service.List(r.Context(), actor(r), rid, cursor)
	if writeError(w, e) {
		return
	}
	items := make([]generated.Object, 0, len(page.Items))
	for _, i := range page.Items {
		items = append(items, view(i))
	}
	transport.WriteJSON(w, 200, generated.ObjectPage{Items: items, NextCursor: page.NextCursor})
}

func (s *objectServer) GetObject(w http.ResponseWriter, r *http.Request, rid, id string) {
	i, e := s.service.Get(r.Context(), actor(r), rid, id)
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

func (s *objectServer) CreateObject(w http.ResponseWriter, r *http.Request, rid string, p generated.CreateObjectParams) {
	var in generated.ObjectCreation
	if !body(w, r, &in) {
		return
	}
	c := object.Content{Name: in.Name, Description: in.Description, Type: in.Type, Quantity: in.Quantity, Unit: in.Unit, SourceDocument: value(in.SourceDocument), SourceRevision: value(in.SourceRevision), SourceItemNo: value(in.SourceItemNo), SourcePages: value(in.SourcePages), MaterialCode: value(in.MaterialCode)}
	i, e := s.service.Create(r.Context(), actor(r), rid, p.IdempotencyKey, c)
	if !writeError(w, e) {
		transport.WriteJSON(w, 200, view(i))
	}
}

func (s *objectServer) UpdateObject(w http.ResponseWriter, r *http.Request, rid, id string, _ generated.UpdateObjectParams) {
	var in generated.ObjectUpdate
	if !body(w, r, &in) {
		return
	}
	u := object.Update{Content: object.Content{Name: in.Name, Description: in.Description, Type: in.Type, Quantity: in.Quantity, Unit: in.Unit, SourceDocument: value(in.SourceDocument), SourceRevision: value(in.SourceRevision), SourceItemNo: value(in.SourceItemNo), SourcePages: value(in.SourcePages), MaterialCode: value(in.MaterialCode)}, Status: string(in.Status), ExpectedVersion: in.ExpectedVersion, Sources: object.SourceUpdate{SourceDocument: in.SourceDocument, SourceRevision: in.SourceRevision, SourceItemNo: in.SourceItemNo, SourcePages: in.SourcePages, MaterialCode: in.MaterialCode}}
	i, e := s.service.Update(r.Context(), actor(r), rid, id, u)
	if !writeError(w, e) {
		transport.WriteJSON(w, 200, view(i))
	}
}
