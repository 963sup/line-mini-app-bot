package repositoryhttp

import (
	"context"
	"errors"
	"net/http"
	"time"

	"github.com/963sup/line-mini-app-bot/internal/business/repository"
	"github.com/963sup/line-mini-app-bot/internal/httpapi"
	"github.com/963sup/line-mini-app-bot/internal/httpapi/generated"
	transport "github.com/963sup/line-mini-app-bot/internal/transport/http"
)

type Authenticator interface {
	Authenticate(http.ResponseWriter, *http.Request) (string, string, bool)
}

func view(r repository.Repository) generated.Repository {
	return generated.Repository{Id: r.ID, GroupId: r.GroupID, Name: r.Name, Status: generated.RepositoryStatus(r.Status), Version: r.Version, CreatedAt: r.CreatedAt, UpdatedAt: r.UpdatedAt}
}

type repositoryServer struct {
	service *repository.Service
	auth    Authenticator
}

var _ httpapi.RepositoryServer = (*repositoryServer)(nil)

func Register(mux *http.ServeMux, service *repository.Service, auth Authenticator) {
	httpapi.RegisterRepositories(mux, &repositoryServer{service: service, auth: auth})
}

func (s *repositoryServer) ListRepositories(w http.ResponseWriter, r *http.Request, params generated.ListRepositoriesParams) {
	cursor := ""
	if params.Cursor != nil {
		cursor = *params.Cursor
	}
	s.read(w, r, "", cursor, false)
}

func (s *repositoryServer) GetRepository(w http.ResponseWriter, r *http.Request, id string) {
	s.read(w, r, id, "", true)
}

func (s *repositoryServer) read(w http.ResponseWriter, r *http.Request, id, cursor string, detail bool) {
	service, auth := s.service, s.auth

	w.Header().Set("Cache-Control", "no-store")
	user, _, ok := auth.Authenticate(w, r)
	if !ok {
		return
	}
	if service == nil {
		transport.WriteAPIError(w, 503, "not_configured", "空間尚未啟用。")
		return
	}
	ctx, cancel := context.WithTimeout(r.Context(), 12*time.Second)
	defer cancel()
	if detail {

		if !repository.ValidID(id) {
			writeError(w, repository.ErrNotFound)
			return
		}
		value, err := service.Get(ctx, user, id)
		if writeError(w, err) {
			return
		}
		transport.WriteJSON(w, 200, view(value))
		return
	}
	values := r.URL.Query()["cursor"]
	if len(values) > 1 || len(cursor) > 1024 {
		writeError(w, repository.ErrCursor)
		return
	}
	page, err := service.List(ctx, user, cursor)
	if writeError(w, err) {
		return
	}
	items := make([]generated.Repository, 0, len(page.Items))
	for _, item := range page.Items {
		items = append(items, view(item))
	}
	transport.WriteJSON(w, 200, generated.RepositoryPage{Items: items, NextCursor: page.NextCursor})
}

func writeError(w http.ResponseWriter, err error) bool {
	if err == nil {
		return false
	}
	switch {
	case errors.Is(err, repository.ErrNotFound):
		transport.WriteAPIError(w, 404, "not_found", "找不到此空間。")
	case errors.Is(err, repository.ErrCursor):
		transport.WriteAPIError(w, 400, "invalid_cursor", "分頁已失效，請重新查詢。")
	default:
		transport.WriteAPIError(w, 503, "unavailable", "暫時無法讀取空間，請重試。")
	}
	return true
}
