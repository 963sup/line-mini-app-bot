package spacehttp

import (
	"context"
	"errors"
	"net/http"
	"time"

	"github.com/963sup/line-mini-app-bot/internal/business/space"
	"github.com/963sup/line-mini-app-bot/internal/httpapi"
	"github.com/963sup/line-mini-app-bot/internal/httpapi/generated"
	transport "github.com/963sup/line-mini-app-bot/internal/transport/http"
)

type Authenticator interface {
	Authenticate(http.ResponseWriter, *http.Request) (string, string, bool)
}

func view(r space.Space) generated.Space {
	return generated.Space{Id: r.ID, GroupId: r.GroupID, Name: r.Name, Status: generated.SpaceStatus(r.Status), Version: r.Version, CreatedAt: r.CreatedAt, UpdatedAt: r.UpdatedAt}
}

type spaceServer struct {
	service *space.Service
	auth    Authenticator
}

var _ httpapi.SpaceServer = (*spaceServer)(nil)

func Register(mux *http.ServeMux, service *space.Service, auth Authenticator) {
	httpapi.RegisterSpaces(mux, &spaceServer{service: service, auth: auth})
}

func (s *spaceServer) ListSpaces(w http.ResponseWriter, r *http.Request, params generated.ListSpacesParams) {
	cursor := ""
	if params.Cursor != nil {
		cursor = *params.Cursor
	}
	s.read(w, r, "", cursor, false)
}

func (s *spaceServer) GetSpace(w http.ResponseWriter, r *http.Request, id string) {
	s.read(w, r, id, "", true)
}

func (s *spaceServer) read(w http.ResponseWriter, r *http.Request, id, cursor string, detail bool) {
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

		if !space.ValidID(id) {
			writeError(w, space.ErrNotFound)
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
		writeError(w, space.ErrCursor)
		return
	}
	page, err := service.List(ctx, user, cursor)
	if writeError(w, err) {
		return
	}
	items := make([]generated.Space, 0, len(page.Items))
	for _, item := range page.Items {
		items = append(items, view(item))
	}
	transport.WriteJSON(w, 200, generated.SpacePage{Items: items, NextCursor: page.NextCursor})
}

func writeError(w http.ResponseWriter, err error) bool {
	if err == nil {
		return false
	}
	switch {
	case errors.Is(err, space.ErrNotFound):
		transport.WriteAPIError(w, 404, "not_found", "找不到此空間。")
	case errors.Is(err, space.ErrCursor):
		transport.WriteAPIError(w, 400, "invalid_cursor", "分頁已失效，請重新查詢。")
	default:
		transport.WriteAPIError(w, 503, "unavailable", "暫時無法讀取空間，請重試。")
	}
	return true
}
