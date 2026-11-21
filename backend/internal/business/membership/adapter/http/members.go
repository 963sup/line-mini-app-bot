// Package membershiphttp exposes Organization People through the existing Identity boundary.
package membershiphttp

import (
	"context"
	"errors"
	"net/http"

	"github.com/963sup/line-mini-app-bot/internal/business/membership/application"
	transport "github.com/963sup/line-mini-app-bot/internal/transport/http"
	"github.com/963sup/line-mini-app-bot/internal/transport/http/api/generated"
)

type Handler struct {
	Service *application.Service
	Auth    interface {
		Authenticate(http.ResponseWriter, *http.Request) (string, string, bool)
	}
}

func (h Handler) serve(w http.ResponseWriter, r *http.Request, operation func(context.Context, string) (any, error)) {
	actor, _, authenticated := h.Auth.Authenticate(w, r)
	if !authenticated {
		return
	}
	if h.Service == nil {
		transport.WriteAPIError(w, 503, "not_configured", "成員服務目前無法使用。")
		return
	}
	result, err := operation(r.Context(), actor)
	switch {
	case errors.Is(err, application.ErrOwnerMustTransfer):
		transport.WriteAPIError(w, 409, "owner_must_transfer", "請先移轉組織所有權。")
	case errors.Is(err, application.ErrDepartureConflict):
		transport.WriteAPIError(w, 409, "conflict", "操作已變更，請重新整理後再試。")
	case errors.Is(err, application.ErrForbidden):
		transport.WriteAPIError(w, 404, "not_found", "找不到組織或成員，或你已無操作權限。")
	case err != nil:
		transport.WriteServiceError(w, 503, "unavailable", "暫時無法完成操作。", err)
	case result == nil:
		w.WriteHeader(http.StatusNoContent)
	default:
		transport.WriteJSON(w, http.StatusOK, result)
	}
}

func (h Handler) ListOrganizationMembers(w http.ResponseWriter, r *http.Request, organizationID string) {
	h.serve(w, r, func(ctx context.Context, actor string) (any, error) {
		roster, err := h.Service.ListMembers(ctx, actor, organizationID)
		if err != nil {
			return nil, err
		}
		result := generated.OrganizationMembers{
			OrganizationName: roster.Name, CanLeave: roster.CanLeave, CanRemove: roster.CanRemove,
			Members: make([]generated.OrganizationMember, 0, len(roster.People)),
		}
		for _, person := range roster.People {
			result.Members = append(result.Members, generated.OrganizationMember{MemberId: person.ID, Owner: person.Owner, Self: person.Self})
		}
		return result, nil
	})
}

func (h Handler) LeaveOrganization(w http.ResponseWriter, r *http.Request, organizationID string, params generated.LeaveOrganizationParams) {
	h.serve(w, r, func(ctx context.Context, actor string) (any, error) {
		return nil, h.Service.EndMembership(ctx, actor, organizationID, "", params.IdempotencyKey)
	})
}

func (h Handler) RemoveOrganizationMember(w http.ResponseWriter, r *http.Request, organizationID, memberID string, params generated.RemoveOrganizationMemberParams) {
	h.serve(w, r, func(ctx context.Context, actor string) (any, error) {
		return nil, h.Service.EndMembership(ctx, actor, organizationID, memberID, params.IdempotencyKey)
	})
}
