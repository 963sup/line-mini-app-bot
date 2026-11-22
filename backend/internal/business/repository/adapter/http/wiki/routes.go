package wikihttp

import (
	repositoryApplicationWiki "github.com/963sup/line-mini-app-bot/internal/business/repository/application/wiki"
	repositoryDomainEntity "github.com/963sup/line-mini-app-bot/internal/business/repository/domain/entity"
	repositoryDomainEntityWiki "github.com/963sup/line-mini-app-bot/internal/business/repository/domain/entity/wiki"
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
	wikiServer struct {
		service *repositoryApplicationWiki.Service
	}
)

var _ httpapi.WikiServer = (*wikiServer)(nil)

func Register(mux *http.ServeMux, service *repositoryApplicationWiki.Service, auth Authenticator) {
	spec, err := httpapi.Spec()
	if err != nil {
		panic(err)
	}
	httpapi.RegisterWiki(mux, &wikiServer{service}, func(next http.Handler) http.Handler {
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
				transport.WriteRequestError(w, 400, "invalid_input", "請求資料不符合Wiki契約。", false)
			},
		})(next)
		return http.NewCrossOriginProtection().Handler(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Cache-Control", "no-store")
			user, name, ok := auth.Authenticate(w, r)
			if !ok {
				return
			}
			if service == nil {
				transport.WriteAPIError(w, 503, "not_configured", "Wiki尚未啟用。")
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
				r.Body = http.MaxBytesReader(w, r.Body, 256<<10)
				if r.Method == http.MethodPost && len(r.Header.Values("Idempotency-Key")) != 1 {
					transport.WriteRequestError(w, 400, "invalid_input", "請提供唯一的重試識別碼。", false)
					return
				}
			}
			ctx, cancel := context.WithTimeout(r.Context(), 12*time.Second)
			defer cancel()
			validate.ServeHTTP(w, r.WithContext(identity.WithUser(ctx, identity.User{ID: user, Name: name})))
		}))
	})
}

func handleError(w http.ResponseWriter, e error) bool {
	if e == nil {
		return false
	}
	switch {
	case errors.Is(e, repositoryDomainEntity.ErrNotFound):
		transport.WriteRequestError(w, 404, "not_found", "找不到知識庫或頁面。", false)
	case errors.Is(e, repositoryDomainEntityWiki.ErrInput):
		transport.WriteRequestError(w, 400, "invalid_input", "請確認標題、內容或分頁參數。", false)
	case errors.Is(e, repositoryDomainEntityWiki.ErrConflict):
		var result generated.APIError
		var reason generated.APIErrorErrorConflictReason
		result.Error.Code = "conflict"
		retrySameRequest := false
		result.Error.RetrySameRequest = &retrySameRequest
		result.Error.Message = "請核對後再儲存。"
		switch {
		case errors.Is(e, repositoryDomainEntityWiki.ErrVersion):
			reason, result.Error.Message = generated.StaleVersion, "頁面已有新版本，請比對後再儲存。"
		case errors.Is(e, repositoryDomainEntityWiki.ErrTitle):
			reason, result.Error.Message = generated.DuplicateTitle, "標題已被使用，請修改標題。"
		case errors.Is(e, repositoryDomainEntityWiki.ErrReplay):
			reason, result.Error.Message = generated.CreationKey, "建立識別碼已使用，請核對原建立請求。"
		case errors.Is(e, repositoryDomainEntityWiki.ErrLimit):
			reason, result.Error.Message = generated.VersionLimit, "頁面版本已達上限。"
		}
		if reason != "" {
			result.Error.ConflictReason = &reason
		}
		transport.WriteJSON(w, http.StatusConflict, result)
	default:
		transport.WriteRequestError(w, 503, "unavailable", "暫時無法存取知識庫。", true)
	}
	return true
}

func page(p repositoryDomainEntityWiki.Page) generated.WikiPage {
	return generated.WikiPage{Id: p.ID, Title: p.Title, Body: p.Body, Version: p.Version, AuthorId: p.AuthorID, UpdatedAt: p.UpdatedAt}
}

func (s *wikiServer) scope(w http.ResponseWriter, r *http.Request, rid string) (repositoryDomainEntity.Repository, bool) {
	repo, e := s.service.Repositories.Get(r.Context(), identity.UserID(r.Context()), rid)
	return repo, !handleError(w, e)
}

func (s *wikiServer) GetWikiPage(w http.ResponseWriter, r *http.Request, rid, id string) {
	repo, ok := s.scope(w, r, rid)
	if !ok {
		return
	}
	p, e := s.service.Store.GetWiki(r.Context(), repo, id)
	if !handleError(w, e) {
		transport.WriteJSON(w, 200, page(p))
	}
}

func (s *wikiServer) ListWikiPages(w http.ResponseWriter, r *http.Request, rid string, params generated.ListWikiPagesParams) {
	s.list(w, r, rid, "", params.Cursor)
}

func (s *wikiServer) ListWikiRevisions(w http.ResponseWriter, r *http.Request, rid, id string, params generated.ListWikiRevisionsParams) {
	s.list(w, r, rid, id, params.Cursor)
}

func (s *wikiServer) list(w http.ResponseWriter, r *http.Request, rid, id string, cursor *string) {
	repo, ok := s.scope(w, r, rid)
	if !ok {
		return
	}
	token := ""
	if cursor != nil {
		token = *cursor
	}
	items, next, e := s.service.Store.ListWiki(r.Context(), repo, identity.UserID(r.Context()), id, token)
	if handleError(w, e) {
		return
	}
	if id == "" {
		out := generated.WikiDirectory{Items: []generated.WikiSummary{}, NextCursor: next}
		for _, p := range items {
			out.Items = append(out.Items, generated.WikiSummary{Id: p.ID, Title: p.Title, Version: p.Version, UpdatedAt: p.UpdatedAt, AuthorId: p.AuthorID})
		}
		transport.WriteJSON(w, 200, out)
		return
	}
	out := generated.WikiHistory{Items: []generated.WikiPage{}, NextCursor: next}
	for _, p := range items {
		out.Items = append(out.Items, page(p))
	}
	transport.WriteJSON(w, 200, out)
}

func (s *wikiServer) CreateWikiPage(w http.ResponseWriter, r *http.Request, rid string, params generated.CreateWikiPageParams) {
	var in generated.WikiCreation
	if !transport.ReadJSON(w, r, &in, "頁面資料無效。") {
		return
	}
	s.save(w, r, rid, "", params.IdempotencyKey, 0, in.Title, in.Body)
}

func (s *wikiServer) UpdateWikiPage(w http.ResponseWriter, r *http.Request, rid, id string) {
	var in generated.WikiUpdate
	if !transport.ReadJSON(w, r, &in, "頁面資料無效。") {
		return
	}
	s.save(w, r, rid, id, "", in.ExpectedVersion, in.Title, in.Body)
}

func (s *wikiServer) save(w http.ResponseWriter, r *http.Request, rid, id, key string, v int64, title, body string) {
	repo, ok := s.scope(w, r, rid)
	if !ok {
		return
	}
	p, e := s.service.Save(r.Context(), repo, identity.UserID(r.Context()), id, key, v, title, body)
	if !handleError(w, e) {
		transport.WriteJSON(w, 200, page(p))
	}
}
