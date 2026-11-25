package repositoryhttp

import (
	repositoryApplication "github.com/963sup/line-mini-app-bot/internal/business/repository/application"
	repositoryDomainEntity "github.com/963sup/line-mini-app-bot/internal/business/repository/domain/entity"
)

import (
	"context"
	"errors"
	"mime"
	"net/http"
	"time"

	transport "github.com/963sup/line-mini-app-bot/internal/transport/http"
	"github.com/963sup/line-mini-app-bot/internal/transport/http/api"
	"github.com/963sup/line-mini-app-bot/internal/transport/http/api/generated"
)

type Authenticator interface {
	Authenticate(http.ResponseWriter, *http.Request) (string, string, bool)
}

func view(r repositoryDomainEntity.Repository) generated.Repository {
	response := generated.Repository{Id: r.ID, OwnerType: generated.RepositoryOwnerType(r.OwnerType), OwnerId: r.OwnerID, Name: r.Name, Status: generated.RepositoryStatus(r.Status), Version: r.Version, CreatedAt: r.CreatedAt, UpdatedAt: r.UpdatedAt}
	if r.Address != nil {
		response.Address = &generated.RepositoryAddress{CountryCode: r.Address.CountryCode, PostalCode: r.Address.PostalCode, AddressLine: r.Address.AddressLine}
	}
	if r.GroupID != "" {
		response.GroupId = &r.GroupID
	}
	return response
}

func (s *repositoryServer) UpdateRepositoryAddress(w http.ResponseWriter, r *http.Request, id string) {
	w.Header().Set("Cache-Control", "no-store")
	user, _, ok := s.auth.Authenticate(w, r)
	if !ok {
		return
	}
	if s.service == nil {
		transport.WriteAPIError(w, 503, "not_configured", "空間尚未啟用。")
		return
	}
	media, _, err := mime.ParseMediaType(r.Header.Get("Content-Type"))
	if err != nil || media != "application/json" {
		transport.WriteAPIError(w, 415, "invalid_content_type", "請使用 JSON。")
		return
	}
	r.Body = http.MaxBytesReader(w, r.Body, 4096)
	var input generated.RepositoryAddressUpdate
	if !transport.ReadJSON(w, r, &input, "地址內容無效。") {
		return
	}
	address, err := repositoryDomainEntity.NewAddress(input.Address.CountryCode, input.Address.PostalCode, input.Address.AddressLine)
	if err != nil || !repositoryDomainEntity.ValidID(id) {
		transport.WriteAPIError(w, 400, "invalid_input", "地址內容無效。")
		return
	}
	ctx, cancel := context.WithTimeout(r.Context(), 12*time.Second)
	defer cancel()
	updated, err := s.service.UpdateAddress(ctx, user, id, input.ExpectedVersion, address, time.Now().UTC())
	if writeError(w, err) {
		return
	}
	transport.WriteJSON(w, 200, view(updated))
}

type repositoryServer struct {
	service *repositoryApplication.Service
	auth    Authenticator
}

var _ httpapi.RepositoryServer = (*repositoryServer)(nil)

func Register(mux *http.ServeMux, service *repositoryApplication.Service, auth Authenticator) {
	httpapi.RegisterRepositories(mux, &repositoryServer{service: service, auth: auth})
}

func (s *repositoryServer) ListRepositories(w http.ResponseWriter, r *http.Request, params generated.ListRepositoriesParams) {
	cursor := ""
	if params.Cursor != nil {
		cursor = *params.Cursor
	}
	s.read(w, r, "", cursor, params.OwnerId, false)
}

func (s *repositoryServer) GetRepository(w http.ResponseWriter, r *http.Request, id string) {
	s.read(w, r, id, "", nil, true)
}

func (s *repositoryServer) read(w http.ResponseWriter, r *http.Request, id, cursor string, owner *string, detail bool) {
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

		if !repositoryDomainEntity.ValidID(id) {
			writeError(w, repositoryDomainEntity.ErrNotFound)
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
		writeError(w, repositoryDomainEntity.ErrCursor)
		return
	}
	ownerID := ""
	if owner != nil {
		if len(r.URL.Query()["ownerId"]) != 1 || !repositoryDomainEntity.ValidOwnerID(*owner) {
			transport.WriteAPIError(w, 400, "invalid_input", "來源擁有者格式無效。")
			return
		}
		ownerID = *owner
	}
	page, err := service.List(ctx, user, cursor, ownerID)
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
	case errors.Is(err, repositoryDomainEntity.ErrNotFound):
		transport.WriteServiceError(w, 404, "not_found", "找不到此空間。", err)
	case errors.Is(err, repositoryDomainEntity.ErrCursor):
		transport.WriteServiceError(w, 400, "invalid_cursor", "分頁已失效，請重新查詢。", err)
	case errors.Is(err, repositoryDomainEntity.ErrInput):
		transport.WriteServiceError(w, 400, "invalid_input", "地址內容無效。", err)
	case errors.Is(err, repositoryDomainEntity.ErrConflict):
		transport.WriteServiceError(w, 409, "conflict", "空間資料已更新，請重新載入。", err)
	default:
		transport.WriteServiceError(w, 503, "unavailable", "暫時無法讀取空間，請重試。", err)
	}
	return true
}
