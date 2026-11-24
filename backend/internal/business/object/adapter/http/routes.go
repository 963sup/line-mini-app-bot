package objecthttp

import (
	objectApplication "github.com/963sup/line-mini-app-bot/internal/business/object/application"
	objectApplicationDto "github.com/963sup/line-mini-app-bot/internal/business/object/application/dto"
	objectDomainEntity "github.com/963sup/line-mini-app-bot/internal/business/object/domain/entity"
)

import (
	"context"
	"errors"
	"mime"
	"net/http"
	"time"

	objectport "github.com/963sup/line-mini-app-bot/internal/business/object/application/port"

	"github.com/963sup/line-mini-app-bot/internal/identity"
	transport "github.com/963sup/line-mini-app-bot/internal/transport/http"
	"github.com/963sup/line-mini-app-bot/internal/transport/http/api"
	"github.com/963sup/line-mini-app-bot/internal/transport/http/api/generated"
	"github.com/getkin/kin-openapi/openapi3filter"
	middleware "github.com/oapi-codegen/nethttp-middleware"
)

const maxCandidateDocumentBytes = 8 << 20

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
				transport.WriteRequestError(w, 400, "invalid_input", "請求資料不符合物件契約。", false)
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
			extraction := r.Pattern == "POST /api/v1/repositories/{repositoryId}/objects/candidates" || r.Pattern == "POST /api/v1/object-template-drafts"
			timeout := 12 * time.Second
			if r.Method == http.MethodPost || r.Method == http.MethodPatch {
				media, _, err := mime.ParseMediaType(r.Header.Get("Content-Type"))
				if err != nil || media != "application/json" {
					transport.WriteAPIError(w, 415, "invalid_content_type", "請使用 JSON。")
					return
				}
				limit := int64(65536)
				if extraction {
					limit = 4*((maxCandidateDocumentBytes+2)/3) + 65536
					timeout = 75 * time.Second
				}
				r.Body = http.MaxBytesReader(w, r.Body, limit)
				if r.Method == http.MethodPost && !extraction && len(r.Header.Values("Idempotency-Key")) != 1 {
					transport.WriteRequestError(w, 400, "invalid_input", "請提供唯一的重試識別碼。", false)
					return
				}
			}
			ctx, cancel := context.WithTimeout(r.Context(), timeout)
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
	var parentID *string
	if i.ParentID != "" {
		parentID = &i.ParentID
	}
	var preset *generated.ObjectWorkConversionDefault
	if value := i.WorkConversionDefault(); value != "" {
		mode := generated.ObjectWorkConversionDefault(value)
		preset = &mode
	}
	return generated.Object{WorkConversionDefault: preset, TemplateId: optional(i.TemplateID), TemplateVersion: templateVersion(i.TemplateVersion), TemplateNodeId: optional(i.TemplateNodeID), ParentId: parentID, RequiresProgressReset: i.RequiresProgressReset(), CanEdit: i.CanEdit(actor), SpaceId: spaceID, CompletedQuantity: completed, Id: i.ID, RepositoryId: i.RepositoryID, Name: i.Name, Description: i.Description, Type: i.Type, Quantity: i.Quantity, Unit: i.Unit, Status: generated.ObjectStatus(i.Status), SourceDocument: i.SourceDocument, SourceRevision: i.SourceRevision, SourceItemNo: i.SourceItemNo, SourcePages: i.SourcePages, MaterialCode: i.MaterialCode, CreatorId: i.CreatorID, Version: i.Version, CreatedAt: i.CreatedAt, UpdatedAt: i.UpdatedAt}
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
		code, httpStatus, message = "invalid_input", 400, "請填寫名稱；數量若有填寫須為正數，追蹤進度時需補齊數量、單位與完成數。"
	case errors.Is(e, objectDomainEntity.ErrConflict):
		code, httpStatus, message = "version_conflict", 409, "物件已更新，請重新讀取。"
	case errors.Is(e, objectDomainEntity.ErrCreationConflict):
		code, httpStatus, message = "idempotency_conflict", 409, "重試識別碼已用於不同內容，請先確認原物件。"
	}
	transport.WriteResourceError(w, httpStatus, code, message, e, code == "unavailable")
	return true
}

func (s *objectServer) ListObjects(w http.ResponseWriter, r *http.Request, rid string, p generated.ListObjectsParams) {
	cursor := ""
	if p.Cursor != nil {
		cursor = *p.Cursor
	}
	page, e := s.service.List(r.Context(), identity.UserID(r.Context()), rid, cursor, value(p.ParentId))
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
	c := objectDomainEntity.Content{ParentID: value(in.ParentId), SpaceID: value(in.SpaceId), CompletedQuantity: value(in.CompletedQuantity), Name: in.Name, Description: value(in.Description), Type: value(in.Type), Quantity: value(in.Quantity), Unit: value(in.Unit), SourceDocument: value(in.SourceDocument), SourceRevision: value(in.SourceRevision), SourceItemNo: value(in.SourceItemNo), SourcePages: value(in.SourcePages), MaterialCode: value(in.MaterialCode)}
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
	actor, _ := identity.FromContext(r.Context())
	u.ActorDisplayName = actor.Name
	i, e := s.service.Update(r.Context(), actor.ID, rid, id, u)
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

func (s *objectServer) ExtractObjectCandidates(w http.ResponseWriter, r *http.Request, rid string, _ generated.ExtractObjectCandidatesParams) {
	var in generated.ObjectCandidateDocument
	if !transport.ReadJSON(w, r, &in, "文件資料無效。") {
		return
	}
	if len(in.Data) == 0 || len(in.Data) > maxCandidateDocumentBytes {
		transport.WriteAPIError(w, 400, "invalid_document", "請選擇 8 MiB 以內的非空文件。")
		return
	}
	result, err := s.service.ExtractCandidates(r.Context(), identity.UserID(r.Context()), rid, objectApplicationDto.Document{Filename: in.Filename, Data: in.Data})
	out := generated.ObjectCandidateResult{Status: "completed", Candidates: []generated.ObjectCandidate{}, RawResponse: string(result.RawResponse)}
	switch {
	case err == nil:
		for _, c := range result.Candidates {
			candidate := generated.ObjectCandidate{Name: c.Name, Quantity: c.Quantity, Unit: c.Unit, Quote: c.Quote, Location: c.Location}
			if len(c.CategoryPath) > 0 {
				candidate.CategoryPath = &c.CategoryPath
			}
			out.Candidates = append(out.Candidates, candidate)
		}
	case errors.Is(err, objectport.ErrPreparationInvalidDocument):
		transport.WriteAPIError(w, 400, "invalid_document", "不支援此文件格式。")
		return
	case errors.Is(err, objectDomainEntity.ErrUnavailable):
		transport.WriteAPIError(w, 503, "not_configured", "文件辨識尚未啟用。")
		return
	case errors.Is(err, objectport.ErrPreparationIncomplete):
		out.Status = "incomplete"
	case errors.Is(err, objectport.ErrPreparationRefused):
		out.Status = "refused"
	case errors.Is(err, objectport.ErrPreparationInvalidOutput):
		out.Status = "invalid_output"
	case errors.Is(err, objectport.ErrPreparationProvider):
		out.Status = "provider_error"
	default:
		writeError(w, err)
		return
	}
	transport.WriteJSON(w, 200, out)
}

func templateVersion(n int64) *int64 {
	if n == 0 {
		return nil
	}
	return &n
}

func (s *objectServer) ListObjectQuantityChanges(w http.ResponseWriter, r *http.Request, rid string, p generated.ListObjectQuantityChangesParams) {
	page, err := s.service.ListQuantityChanges(r.Context(), identity.UserID(r.Context()), rid, value(p.Cursor))
	if writeError(w, err) {
		return
	}
	items := make([]generated.ObjectQuantityChange, 0, len(page.Items))
	for _, c := range page.Items {
		items = append(items, generated.ObjectQuantityChange{ObjectId: c.ObjectID, Name: c.Name, Version: c.Version, Actor: c.Actor, At: c.At, BeforeQuantity: c.BeforeQuantity, BeforeUnit: c.BeforeUnit, AfterQuantity: c.AfterQuantity, AfterUnit: c.AfterUnit})
	}
	transport.WriteJSON(w, 200, generated.ObjectQuantityChangePage{Items: items, NextCursor: page.NextCursor})
}
