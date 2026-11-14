package objectimporthttp

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"mime"
	"net/http"
	"time"

	"github.com/963sup/line-mini-app-bot/internal/business/objectimport"
	"github.com/963sup/line-mini-app-bot/internal/business/objectimport/parsing"
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
	importServer struct {
		service *objectimport.Service
		parser  *parsing.Service
	}
)

var _ httpapi.ObjectImportServer = (*importServer)(nil)

func Register(mux *http.ServeMux, service *objectimport.Service, auth Authenticator, parsers ...*parsing.Service) {
	spec, err := httpapi.Spec()
	if err != nil {
		panic(err)
	}
	var parser *parsing.Service
	if len(parsers) > 0 {
		parser = parsers[0]
	}
	httpapi.RegisterObjectImports(mux, &importServer{service: service, parser: parser}, func(next http.Handler) http.Handler {
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
				r.Body = http.MaxBytesReader(w, r.Body, 512*1024)
				if len(r.Header.Values("Idempotency-Key")) > 1 {
					transport.WriteAPIError(w, 400, "invalid_input", "請提供唯一的重試識別碼。")
					return
				}
			}
			timeout := 12 * time.Second
			if r.Pattern == "POST /api/v1/repositories/{repositoryId}/object-imports/parse" {
				timeout = 85 * time.Second
			}
			ctx, cancel := context.WithTimeout(r.Context(), timeout)
			defer cancel()
			validate.ServeHTTP(w, r.WithContext(identity.WithUser(ctx, identity.User{ID: user, Name: name})))
		}))
	})
}

func view(j objectimport.Job) generated.ObjectImport {
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
func actor(r *http.Request) string { u, _ := identity.FromContext(r.Context()); return u.ID }
func writeError(w http.ResponseWriter, e error) bool {
	if e == nil {
		return false
	}
	code, httpStatus, message := "unavailable", 503, "匯入結果尚未確認，請讀取原草稿並沿用原審核版本重試；建立草稿請沿用原識別碼。"
	switch {
	case errors.Is(e, objectimport.ErrNotFound):
		code, httpStatus, message = "not_found", 404, "找不到此空間或匯入草稿。"
	case errors.Is(e, objectimport.ErrForbidden):
		code, httpStatus, message = "forbidden", 403, "只有建立者可以操作此匯入草稿。"
	case errors.Is(e, objectimport.ErrInput):
		code, httpStatus, message = "invalid_input", 400, "候選、來源引用或大小不符合草稿規則。"
	case errors.Is(e, objectimport.ErrConflict):
		code, httpStatus, message = "version_conflict", 409, "草稿版本或狀態不符，請重新讀取。"
	case errors.Is(e, objectimport.ErrCreationConflict):
		code, httpStatus, message = "idempotency_conflict", 409, "識別碼已用於不同草稿，請先確認原結果。"
	}
	transport.WriteAPIError(w, httpStatus, code, message)
	return true
}

func body(w http.ResponseWriter, r *http.Request, target any) bool {
	d := json.NewDecoder(r.Body)
	d.DisallowUnknownFields()
	if d.Decode(target) != nil || d.Decode(new(any)) != io.EOF {
		transport.WriteAPIError(w, 400, "invalid_input", "草稿 JSON 無效。")
		return false
	}
	return true
}

func (s *importServer) CreateObjectImport(w http.ResponseWriter, r *http.Request, rid string, p generated.CreateObjectImportParams) {
	var in generated.ObjectImportCreation
	if !body(w, r, &in) {
		return
	}
	c := objectimport.Creation{SourceDocument: in.SourceDocument, SourceRevision: in.SourceRevision, Pages: []objectimport.Page{}, Rows: []objectimport.Candidate{}}
	for _, page := range in.Pages {
		c.Pages = append(c.Pages, objectimport.Page{Number: page.Number, Text: page.Text})
	}
	for _, row := range in.Rows {
		candidate := objectimport.Candidate{Fields: objectimport.Fields{Name: row.Name, Description: row.Description, Type: row.Type, Quantity: row.Quantity, Unit: row.Unit, MaterialCode: row.MaterialCode}, SourceItemNo: row.SourceItemNo}
		for _, e := range row.Evidence {
			candidate.Evidence = append(candidate.Evidence, objectimport.Evidence{Page: e.Page, Quote: e.Quote})
		}
		c.Rows = append(c.Rows, candidate)
	}
	j, e := s.service.Create(r.Context(), actor(r), rid, p.IdempotencyKey, c)
	if !writeError(w, e) {
		transport.WriteJSON(w, 200, view(j))
	}
}

func (s *importServer) GetObjectImport(w http.ResponseWriter, r *http.Request, rid, id string) {
	j, e := s.service.Get(r.Context(), actor(r), rid, id)
	if !writeError(w, e) {
		transport.WriteJSON(w, 200, view(j))
	}
}

func (s *importServer) ReviewObjectImport(w http.ResponseWriter, r *http.Request, rid, id string, _ generated.ReviewObjectImportParams) {
	var in generated.ObjectImportReview
	if !body(w, r, &in) {
		return
	}
	input := objectimport.Review{ExpectedVersion: in.ExpectedVersion, Rows: []objectimport.ReviewedRow{}}
	for _, row := range in.Rows {
		input.Rows = append(input.Rows, objectimport.ReviewedRow{ID: row.Id, Fields: objectimport.Fields{Name: row.Name, Description: row.Description, Type: row.Type, Quantity: row.Quantity, Unit: row.Unit, MaterialCode: row.MaterialCode}})
	}
	j, e := s.service.Review(r.Context(), actor(r), rid, id, input)
	if !writeError(w, e) {
		transport.WriteJSON(w, 200, view(j))
	}
}

func (s *importServer) ConfirmObjectImport(w http.ResponseWriter, r *http.Request, rid, id string, _ generated.ConfirmObjectImportParams) {
	var in generated.ObjectImportConfirmation
	if !body(w, r, &in) {
		return
	}
	j, e := s.service.Confirm(r.Context(), actor(r), rid, id, in.ExpectedVersion)
	if !writeError(w, e) {
		transport.WriteJSON(w, 200, view(j))
	}
}

func (s *importServer) ParseObjectImport(w http.ResponseWriter, r *http.Request, rid string, _ generated.ParseObjectImportParams) {
	var in generated.ObjectImportParsing
	if !body(w, r, &in) {
		return
	}
	source := objectimport.Source{SourceDocument: in.SourceDocument, SourceRevision: in.SourceRevision}
	for _, page := range in.Pages {
		source.Pages = append(source.Pages, objectimport.Page{Number: page.Number, Text: page.Text})
	}
	result, e := s.parser.Parse(r.Context(), actor(r), rid, parsing.Input{Source: source, Consent: bool(in.Consent)})
	if e != nil {
		switch {
		case errors.Is(e, parsing.ErrNotConfigured):
			transport.WriteAPIError(w, 503, "not_configured", "工程項目解析尚未設定模型，原草稿與工程項目不會改變。")
		case errors.Is(e, parsing.ErrModel), errors.Is(e, parsing.ErrOutput), errors.Is(e, context.Canceled), errors.Is(e, context.DeadlineExceeded):
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
