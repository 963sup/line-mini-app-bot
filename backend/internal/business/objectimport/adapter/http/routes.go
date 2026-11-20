package objectimporthttp

import (
	objectimportApplication "github.com/963sup/line-mini-app-bot/internal/business/objectimport/application"
	objectimportApplicationAutomatic "github.com/963sup/line-mini-app-bot/internal/business/objectimport/application/automatic"
	objectimportApplicationParsing "github.com/963sup/line-mini-app-bot/internal/business/objectimport/application/parsing"
	objectimportDomainEntity "github.com/963sup/line-mini-app-bot/internal/business/objectimport/domain/entity"
	objectimportDomainEntityParsing "github.com/963sup/line-mini-app-bot/internal/business/objectimport/domain/entity/parsing"
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
	candidateParser interface {
		Parse(context.Context, string, string, objectimportDomainEntityParsing.Input) (objectimportDomainEntityParsing.Result, error)
	}
	Authenticator interface {
		Authenticate(http.ResponseWriter, *http.Request) (string, string, bool)
	}
	importServer struct {
		service   *objectimportApplication.Service
		parser    candidateParser
		automatic *objectimportApplicationAutomatic.Service
	}
)

var _ httpapi.ObjectImportServer = (*importServer)(nil)

func Register(mux *http.ServeMux, service *objectimportApplication.Service, auth Authenticator, auto *objectimportApplicationAutomatic.Service, parser candidateParser) {
	if parser == nil {
		parser = (*objectimportApplicationParsing.Service)(nil)
	}
	spec, err := httpapi.Spec()
	if err != nil {
		panic(err)
	}
	httpapi.RegisterObjectImports(mux, &importServer{service: service, parser: parser, automatic: auto}, func(next http.Handler) http.Handler {
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
				transport.WriteAPIError(w, 400, "invalid_input", "請求資料不符合匯入草稿契約。")
			},
		})(next)
		return http.NewCrossOriginProtection().Handler(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Cache-Control", "no-store")
			user, name, ok := auth.Authenticate(w, r)
			if !ok {
				return
			}
			if service == nil {
				transport.WriteAPIError(w, 503, "not_configured", "匯入草稿尚未啟用。")
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
				limit := int64(512 * 1024)
				if r.Pattern == "POST /api/v1/repositories/{repositoryId}/object-import-runs" {
					limit = 14 << 20
				}
				r.Body = http.MaxBytesReader(w, r.Body, limit)
				if len(r.Header.Values("Idempotency-Key")) > 1 {
					transport.WriteAPIError(w, 400, "invalid_input", "請提供唯一的重試識別碼。")
					return
				}
			}
			timeout := 12 * time.Second
			if r.Pattern == "POST /api/v1/repositories/{repositoryId}/object-imports/parse" || r.Pattern == "POST /api/v1/repositories/{repositoryId}/object-import-runs" {
				timeout = 85 * time.Second
			}
			ctx, cancel := context.WithTimeout(r.Context(), timeout)
			defer cancel()
			validate.ServeHTTP(w, r.WithContext(identity.WithUser(ctx, identity.User{ID: user, Name: name})))
		}))
	})
}

func view(j objectimportDomainEntity.Job) generated.ObjectImport {
	rows := make([]generated.ObjectImportRow, 0, len(j.Rows))
	for _, r := range j.Rows {
		evidence := make([]generated.ObjectImportEvidence, 0, len(r.Evidence))
		for _, e := range r.Evidence {
			evidence = append(evidence, generated.ObjectImportEvidence{Page: e.Page, Quote: e.Quote})
		}
		rows = append(rows, generated.ObjectImportRow{Id: r.ID, Name: r.Name, Description: r.Description, Type: r.Type, Quantity: r.Quantity, Unit: r.Unit, MaterialCode: r.MaterialCode, SourceItemNo: r.SourceItemNo, Evidence: evidence, ResultObjectId: r.ResultObjectID})
	}
	return generated.ObjectImport{Id: j.ID, RepositoryId: j.RepositoryID, CreatorId: j.CreatorID, Status: generated.ObjectImportStatus(j.Status), Version: j.Version, ConfirmedFromReviewVersion: j.ConfirmedFromReviewVersion, SourceDocument: j.SourceDocument, SourceRevision: j.SourceRevision, SourceHash: j.SourceHash, Rows: rows, CreatedAt: j.CreatedAt, UpdatedAt: j.UpdatedAt, ReviewedAt: j.ReviewedAt}
}

func writeError(w http.ResponseWriter, e error) bool {
	if e == nil {
		return false
	}
	code, httpStatus, message := "unavailable", 503, "匯入結果尚未確認，請讀取原草稿並沿用原審核版本重試；建立草稿請沿用原識別碼。"
	switch {
	case errors.Is(e, objectimportDomainEntity.ErrNotFound):
		code, httpStatus, message = "not_found", 404, "找不到此空間或匯入草稿。"
	case errors.Is(e, objectimportDomainEntity.ErrForbidden):
		code, httpStatus, message = "forbidden", 403, "只有建立者可以操作此匯入草稿。"
	case errors.Is(e, objectimportDomainEntity.ErrInput):
		code, httpStatus, message = "invalid_input", 400, "候選、來源引用或大小不符合草稿規則。"
	case errors.Is(e, objectimportDomainEntity.ErrConflict):
		code, httpStatus, message = "version_conflict", 409, "草稿版本或狀態不符，請重新讀取。"
	case errors.Is(e, objectimportDomainEntity.ErrCreationConflict):
		code, httpStatus, message = "idempotency_conflict", 409, "識別碼已用於不同草稿，請先確認原結果。"
	}
	transport.WriteServiceError(w, httpStatus, code, message, e)
	return true
}

func (s *importServer) CreateObjectImport(w http.ResponseWriter, r *http.Request, rid string, p generated.CreateObjectImportParams) {
	var in generated.ObjectImportCreation
	if !transport.ReadJSON(w, r, &in, "草稿 JSON 無效。") {
		return
	}
	c := objectimportDomainEntity.Creation{SourceDocument: in.SourceDocument, SourceRevision: in.SourceRevision, Pages: []objectimportDomainEntity.Page{}, Rows: []objectimportDomainEntity.Candidate{}}
	for _, page := range in.Pages {
		c.Pages = append(c.Pages, objectimportDomainEntity.Page{Number: page.Number, Text: page.Text})
	}
	for _, row := range in.Rows {
		candidate := objectimportDomainEntity.Candidate{Fields: objectimportDomainEntity.Fields{Name: row.Name, Description: row.Description, Type: row.Type, Quantity: row.Quantity, Unit: row.Unit, MaterialCode: row.MaterialCode}, SourceItemNo: row.SourceItemNo}
		for _, e := range row.Evidence {
			candidate.Evidence = append(candidate.Evidence, objectimportDomainEntity.Evidence{Page: e.Page, Quote: e.Quote})
		}
		c.Rows = append(c.Rows, candidate)
	}
	j, e := s.service.Create(r.Context(), identity.UserID(r.Context()), rid, p.IdempotencyKey, c)
	if !writeError(w, e) {
		transport.WriteJSON(w, 200, view(j))
	}
}

func (s *importServer) GetObjectImport(w http.ResponseWriter, r *http.Request, rid, id string) {
	j, e := s.service.Get(r.Context(), identity.UserID(r.Context()), rid, id)
	if !writeError(w, e) {
		transport.WriteJSON(w, 200, view(j))
	}
}

func (s *importServer) ReviewObjectImport(w http.ResponseWriter, r *http.Request, rid, id string, _ generated.ReviewObjectImportParams) {
	var in generated.ObjectImportReview
	if !transport.ReadJSON(w, r, &in, "草稿 JSON 無效。") {
		return
	}
	input := objectimportDomainEntity.Review{ExpectedVersion: in.ExpectedVersion, Rows: []objectimportDomainEntity.ReviewedRow{}}
	for _, row := range in.Rows {
		input.Rows = append(input.Rows, objectimportDomainEntity.ReviewedRow{ID: row.Id, Fields: objectimportDomainEntity.Fields{Name: row.Name, Description: row.Description, Type: row.Type, Quantity: row.Quantity, Unit: row.Unit, MaterialCode: row.MaterialCode}})
	}
	j, e := s.service.Review(r.Context(), identity.UserID(r.Context()), rid, id, input)
	if !writeError(w, e) {
		transport.WriteJSON(w, 200, view(j))
	}
}

func (s *importServer) ConfirmObjectImport(w http.ResponseWriter, r *http.Request, rid, id string, _ generated.ConfirmObjectImportParams) {
	var in generated.ObjectImportConfirmation
	if !transport.ReadJSON(w, r, &in, "草稿 JSON 無效。") {
		return
	}
	j, e := s.service.Confirm(r.Context(), identity.UserID(r.Context()), rid, id, in.ExpectedVersion)
	if !writeError(w, e) {
		transport.WriteJSON(w, 200, view(j))
	}
}

func (s *importServer) ParseObjectImport(w http.ResponseWriter, r *http.Request, rid string, _ generated.ParseObjectImportParams) {
	var in generated.ObjectImportParsing
	if !transport.ReadJSON(w, r, &in, "草稿 JSON 無效。") {
		return
	}
	source := objectimportDomainEntity.Source{SourceDocument: in.SourceDocument, SourceRevision: in.SourceRevision}
	for _, page := range in.Pages {
		source.Pages = append(source.Pages, objectimportDomainEntity.Page{Number: page.Number, Text: page.Text})
	}
	result, e := s.parser.Parse(r.Context(), identity.UserID(r.Context()), rid, objectimportDomainEntityParsing.Input{Source: source, Consent: bool(in.Consent)})
	if e != nil {
		switch {
		case errors.Is(e, objectimportDomainEntityParsing.ErrNotConfigured):
			transport.WriteAPIError(w, 503, "not_configured", "物件解析尚未設定模型，原草稿與物件不會改變。")
		case errors.Is(e, objectimportDomainEntityParsing.ErrModel), errors.Is(e, objectimportDomainEntityParsing.ErrOutput), errors.Is(e, context.Canceled), errors.Is(e, context.DeadlineExceeded):
			transport.WriteAPIError(w, 503, "parse_failed", "模型未產生可驗證的候選，請檢查選取文字或稍後重試。")
		default:
			writeError(w, e)
		}
		return
	}
	rows := make([]generated.ObjectImportCandidate, 0, len(result.Rows))
	for _, r := range result.Rows {
		proof := make([]generated.ObjectImportEvidence, 0, len(r.Evidence))
		for _, e := range r.Evidence {
			proof = append(proof, generated.ObjectImportEvidence{Page: e.Page, Quote: e.Quote})
		}
		rows = append(rows, generated.ObjectImportCandidate{Name: r.Name, Description: r.Description, Type: r.Type, Quantity: r.Quantity, Unit: r.Unit, MaterialCode: r.MaterialCode, SourceItemNo: r.SourceItemNo, Evidence: proof})
	}
	transport.WriteJSON(w, 200, generated.ObjectImportParsed{Rows: rows, SourceHash: result.SourceHash})
}
