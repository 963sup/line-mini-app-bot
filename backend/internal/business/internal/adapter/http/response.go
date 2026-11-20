package orghttp

import (
	organizationDomainEntity "github.com/963sup/line-mini-app-bot/internal/business/organization/domain/entity"
)

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"

	transport "github.com/963sup/line-mini-app-bot/internal/transport/http"
)

type Authenticator interface {
	Authenticate(http.ResponseWriter, *http.Request) (string, string, bool)
}

func Serve(w http.ResponseWriter, r *http.Request, auth Authenticator, configured bool, op func(context.Context, string) (any, error)) {
	user, _, ok := auth.Authenticate(w, r)
	if !ok {
		return
	}
	if !configured {
		transport.WriteAPIError(w, 503, "not_configured", "組織服務尚未設定。")
		return
	}
	result, err := op(r.Context(), user)
	switch {
	case errors.Is(err, organizationDomainEntity.ErrForbidden):
		transport.WriteAPIError(w, 403, "forbidden", "帳號尚無企業開通資格。")
	case errors.Is(err, organizationDomainEntity.ErrInvalid):
		transport.WriteAPIError(w, 400, "invalid_input", "請求資料無效。")
	case errors.Is(err, organizationDomainEntity.ErrNotFound):
		transport.WriteAPIError(w, 404, "not_found", "找不到資源。")
	case errors.Is(err, organizationDomainEntity.ErrConflict):
		transport.WriteAPIError(w, 409, "conflict", "資源已有其他管理歸屬。")
	case err != nil:
		transport.WriteServiceError(w, 503, "unavailable", "暫時無法完成操作。", err)
	case result == nil:
		w.WriteHeader(204)
	default:
		transport.WriteJSON(w, 200, result)
	}
}

func Decode(r *http.Request, w http.ResponseWriter, v any) error {
	dec := json.NewDecoder(http.MaxBytesReader(w, r.Body, 4096))
	dec.DisallowUnknownFields()
	if dec.Decode(v) != nil || dec.Decode(&struct{}{}) != io.EOF {
		return organizationDomainEntity.ErrInvalid
	}
	return nil
}
