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
		transport.WriteRequestError(w, 503, "not_configured", "組織服務尚未設定。", false)
		return
	}
	result, err := op(r.Context(), user)
	switch {
	case errors.Is(err, organizationDomainEntity.ErrForbidden):
		transport.WriteRequestError(w, 403, "forbidden", "帳號尚無企業開通資格。", false)
	case errors.Is(err, organizationDomainEntity.ErrInvalid):
		transport.WriteRequestError(w, 400, "invalid_input", "請求資料無效。", false)
	case errors.Is(err, organizationDomainEntity.ErrNotFound):
		transport.WriteRequestError(w, 404, "not_found", "找不到資源。", false)
	case errors.Is(err, organizationDomainEntity.ErrConflict):
		transport.WriteRequestError(w, 409, "conflict", "資源已有其他管理歸屬。", false)
	case err != nil:
		transport.WriteResourceError(w, 503, "unavailable", "暫時無法完成操作。", err, true)
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
