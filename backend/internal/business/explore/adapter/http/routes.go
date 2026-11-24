package explorehttp

import (
	"context"
	"errors"
	"net/http"
	"time"

	"github.com/963sup/line-mini-app-bot/internal/business/explore/application/query"
	transport "github.com/963sup/line-mini-app-bot/internal/transport/http"
	httpapi "github.com/963sup/line-mini-app-bot/internal/transport/http/api"
	"github.com/963sup/line-mini-app-bot/internal/transport/http/api/generated"
)

type Authenticator interface {
	Authenticate(http.ResponseWriter, *http.Request) (string, string, bool)
}

type server struct {
	service query.Service
	auth    Authenticator
}

var _ httpapi.ExploreServer = (*server)(nil)

func Register(mux *http.ServeMux, service query.Service, auth Authenticator) {
	httpapi.RegisterExplore(mux, &server{service: service, auth: auth})
}

func (s *server) SearchExplore(w http.ResponseWriter, r *http.Request, params generated.SearchExploreParams) {
	w.Header().Set("Cache-Control", "no-store")
	actor, _, ok := s.auth.Authenticate(w, r)
	if !ok {
		return
	}
	cursor := ""
	if params.Cursor != nil {
		cursor = *params.Cursor
	}
	ctx, cancel := context.WithTimeout(r.Context(), 12*time.Second)
	defer cancel()
	page, err := s.service.Search(ctx, actor, string(params.Type), params.Q, cursor)
	if err != nil {
		if errors.Is(err, query.ErrInput) {
			transport.WriteServiceError(w, http.StatusBadRequest, "invalid_request", "搜尋條件或分頁游標無效。", err)
			return
		}
		transport.WriteServiceError(w, http.StatusServiceUnavailable, "unavailable", "暫時無法搜尋，請重試。", err)
		return
	}
	items := make([]generated.ExploreItem, 0, len(page.Items))
	for _, item := range page.Items {
		items = append(items, generated.ExploreItem{
			Type: generated.ExploreItemTypeRepository, Id: item.ID, DisplayName: item.DisplayName,
		})
	}
	transport.WriteJSON(w, http.StatusOK, generated.ExplorePage{Items: items, NextCursor: page.NextCursor})
}
