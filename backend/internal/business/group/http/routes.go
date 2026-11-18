package grouphttp

import (
	"context"
	"errors"
	"net/http"
	"time"

	groupdomain "github.com/963sup/line-mini-app-bot/internal/business/group"
	"github.com/963sup/line-mini-app-bot/internal/httpapi"
	"github.com/963sup/line-mini-app-bot/internal/httpapi/generated"
	transport "github.com/963sup/line-mini-app-bot/internal/transport/http"
)

type Authenticator interface {
	Authenticate(http.ResponseWriter, *http.Request) (string, string, bool)
}

func view(r groupdomain.Group) generated.Group {
	return generated.Group{Id: r.ID, GroupId: r.GroupID, Name: r.Name, Status: generated.GroupStatus(r.Status), Version: r.Version, CreatedAt: r.CreatedAt, UpdatedAt: r.UpdatedAt}
}

type groupServer struct {
	service *groupdomain.Service
	auth    Authenticator
}

var _ httpapi.GroupServer = (*groupServer)(nil)

func Register(mux *http.ServeMux, service *groupdomain.Service, auth Authenticator) {
	httpapi.RegisterGroups(mux, &groupServer{service: service, auth: auth})
}

func (s *groupServer) ListGroups(w http.ResponseWriter, r *http.Request, params generated.ListGroupsParams) {
	cursor := ""
	if params.Cursor != nil {
		cursor = *params.Cursor
	}
	s.read(w, r, "", cursor, false)
}

func (s *groupServer) GetGroup(w http.ResponseWriter, r *http.Request, id string) {
	s.read(w, r, id, "", true)
}

func (s *groupServer) read(w http.ResponseWriter, r *http.Request, id, cursor string, detail bool) {
	service, auth := s.service, s.auth

	w.Header().Set("Cache-Control", "no-store")
	user, _, ok := auth.Authenticate(w, r)
	if !ok {
		return
	}
	if service == nil {
		transport.WriteAPIError(w, 503, "not_configured", "群組尚未啟用。")
		return
	}
	ctx, cancel := context.WithTimeout(r.Context(), 12*time.Second)
	defer cancel()
	if detail {

		if !groupdomain.ValidID(id) {
			writeError(w, groupdomain.ErrNotFound)
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
		writeError(w, groupdomain.ErrCursor)
		return
	}
	page, err := service.List(ctx, user, cursor)
	if writeError(w, err) {
		return
	}
	items := make([]generated.Group, 0, len(page.Items))
	for _, item := range page.Items {
		items = append(items, view(item))
	}
	transport.WriteJSON(w, 200, generated.EngineeringGroupPage{Items: items, NextCursor: page.NextCursor})
}

func writeError(w http.ResponseWriter, err error) bool {
	if err == nil {
		return false
	}
	switch {
	case errors.Is(err, groupdomain.ErrNotFound):
		transport.WriteAPIError(w, 404, "not_found", "找不到此群組。")
	case errors.Is(err, groupdomain.ErrCursor):
		transport.WriteAPIError(w, 400, "invalid_cursor", "分頁已失效，請重新查詢。")
	default:
		transport.WriteAPIError(w, 503, "unavailable", "暫時無法讀取群組，請重試。")
	}
	return true
}
