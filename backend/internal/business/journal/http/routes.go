package journalhttp

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"mime"
	"net/http"
	"time"

	types "github.com/oapi-codegen/runtime/types"

	feature "github.com/963sup/line-mini-app-bot/internal/business/journal"
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
	journalServer struct{ service feature.Operations }
)

var _ httpapi.JournalServer = (*journalServer)(nil)

func Register(mux *http.ServeMux, service feature.Operations, auth Authenticator) {
	spec, err := httpapi.Spec()
	if err != nil {
		panic(err)
	}
	httpapi.RegisterJournals(mux, &journalServer{service}, func(next http.Handler) http.Handler {
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
				transport.WriteAPIError(w, 400, "invalid_input", "請求資料不符合日誌契約。")
			},
		})(next)
		return http.NewCrossOriginProtection().Handler(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Cache-Control", "no-store")
			user, name, ok := auth.Authenticate(w, r)
			if !ok {
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

func replyError(w http.ResponseWriter, err error) bool {
	if err == nil {
		return false
	}
	switch {
	case errors.Is(err, feature.ErrInput):
		transport.WriteServiceError(w, 400, "invalid_input", "日誌資料無效。", err)
	case errors.Is(err, feature.ErrNotFound):
		transport.WriteServiceError(w, 404, "not_found", "找不到日誌或關聯來源。", err)
	case errors.Is(err, feature.ErrConflict):
		transport.WriteServiceError(w, 409, "idempotency_conflict", "重試識別碼已用於不同內容。", err)
	default:
		transport.WriteServiceError(w, 503, "unavailable", "日誌服務暫時無法使用。", err)
	}
	return true
}

func (s *journalServer) enabled(w http.ResponseWriter) bool {
	if s.service != nil {
		return true
	}
	transport.WriteAPIError(w, 503, "not_configured", "群組日誌尚未啟用。")
	return false
}

func source(r *generated.JournalReference) *feature.Reference {
	if r == nil {
		return nil
	}
	return &feature.Reference{Kind: string(r.Type), ID: r.Id}
}

func refView(r *feature.Reference) *generated.JournalReference {
	if r == nil {
		return nil
	}
	return &generated.JournalReference{Type: generated.JournalReferenceType(r.Kind), Id: r.ID}
}
func names(values []string) []string { return append([]string{}, values...) }
func view(e feature.Entry) generated.Journal {
	j := generated.Journal{Id: e.ID, CreatorId: e.CreatorID, CreatedAt: e.CreatedAt, GroupId: e.GroupID, WorkDate: types.Date{Time: e.WorkDate}, Personnel: generated.JournalPersonnel{Management: names(e.Personnel.Management), Control: names(e.Personnel.Control), Workers: names(e.Personnel.Workers)}, TodayItems: []generated.JournalTodayItem{}, TomorrowItems: []generated.JournalTomorrowItem{}}
	for _, i := range e.Today {
		j.TodayItems = append(j.TodayItems, generated.JournalTodayItem{Name: i.Name, CompletionPercent: i.Percent, Reference: refView(i.Reference)})
	}
	for _, i := range e.Tomorrow {
		j.TomorrowItems = append(j.TomorrowItems, generated.JournalTomorrowItem{Name: i.Name, ExpectedCompletionPercent: i.Percent, Reference: refView(i.Reference)})
	}
	return j
}

func (s *journalServer) CreateJournal(w http.ResponseWriter, r *http.Request, p generated.CreateJournalParams) {
	var b generated.JournalCreation
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()
	if decoder.Decode(&b) != nil || decoder.Decode(new(any)) != io.EOF || len(b.TodayItems)+len(b.TomorrowItems) == 0 {
		replyError(w, feature.ErrInput)
		return
	}
	if !s.enabled(w) {
		return
	}
	d := feature.Draft{GroupID: b.GroupId, WorkDate: b.WorkDate.Time, Personnel: feature.Personnel{Control: b.Personnel.Control, Management: b.Personnel.Management, Workers: b.Personnel.Workers}}
	for _, i := range b.TodayItems {
		d.Today = append(d.Today, feature.Work{Name: i.Name, Percent: i.CompletionPercent, Reference: source(i.Reference)})
	}
	for _, i := range b.TomorrowItems {
		d.Tomorrow = append(d.Tomorrow, feature.Work{Name: i.Name, Percent: i.ExpectedCompletionPercent, Reference: source(i.Reference)})
	}
	e, err := s.service.Create(r.Context(), identity.UserID(r.Context()), p.IdempotencyKey, d)
	if !replyError(w, err) {
		transport.WriteJSON(w, 201, view(e))
	}
}

func (s *journalServer) GetJournal(w http.ResponseWriter, r *http.Request, id string) {
	if !s.enabled(w) {
		return
	}
	e, err := s.service.Get(r.Context(), identity.UserID(r.Context()), id)
	if !replyError(w, err) {
		transport.WriteJSON(w, 200, view(e))
	}
}

func (s *journalServer) ListJournals(w http.ResponseWriter, r *http.Request, p generated.ListJournalsParams) {
	if !s.enabled(w) {
		return
	}
	var f feature.Filter
	if p.GroupId != nil {
		f.GroupID = *p.GroupId
	}
	if p.WorkDate != nil {
		f.WorkDate = p.WorkDate.String()
	}
	if p.Cursor != nil {
		f.Cursor = *p.Cursor
	}
	page, err := s.service.List(r.Context(), identity.UserID(r.Context()), f)
	if replyError(w, err) {
		return
	}
	items := make([]generated.Journal, 0, len(page.Items))
	for _, e := range page.Items {
		items = append(items, view(e))
	}
	transport.WriteJSON(w, 200, generated.JournalPage{Items: items, NextCursor: page.NextCursor})
}
