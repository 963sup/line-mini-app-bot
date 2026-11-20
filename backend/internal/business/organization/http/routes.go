package organizationhttp

import (
	"context"
	"encoding/json"
	"errors"
	"github.com/963sup/line-mini-app-bot/internal/business/enterprise"
	"github.com/963sup/line-mini-app-bot/internal/business/team"
	"io"
	"net/http"

	"github.com/963sup/line-mini-app-bot/internal/business/organization"
	"github.com/963sup/line-mini-app-bot/internal/httpapi/generated"
	transport "github.com/963sup/line-mini-app-bot/internal/transport/http"
)

type EnterpriseStore interface {
	CreateEnterprise(context.Context, string, string) (*enterprise.Link, error)
	Enterprise(context.Context, string, string) (*enterprise.Detail, error)
	UpdateEnterprise(context.Context, string, string, string, bool) (*enterprise.Link, error)
}

type OrganizationStore interface {
	Context(ctx context.Context, user string, preference string, write bool) (*organization.Context, error)
	Create(ctx context.Context, user string, name string) (*organization.Organization, error)
	Transfer(ctx context.Context, user string, organizationId, target string) (*organization.Organization, error)
}

type TeamStore interface {
	Manage(ctx context.Context, user string, teamId, scope, managementId string) error
	Assign(ctx context.Context, user string, id, teamId string, add bool) error
	List(ctx context.Context, user string, organizationId string) (*team.Listing, error)
}

type Handler struct {
	EnterpriseStore   EnterpriseStore
	OrganizationStore OrganizationStore
	TeamStore         TeamStore
	Auth              interface {
		Authenticate(http.ResponseWriter, *http.Request) (string, string, bool)
	}
}

func (h Handler) serve(w http.ResponseWriter, r *http.Request, configured bool, op func(context.Context, string) (any, error)) {
	user, _, ok := h.Auth.Authenticate(w, r)
	if !ok {
		return
	}
	if !configured {
		transport.WriteAPIError(w, 503, "not_configured", "組織服務尚未設定。")
		return
	}
	result, err := op(r.Context(), user)
	switch {
	case errors.Is(err, organization.ErrForbidden):
		transport.WriteAPIError(w, 403, "forbidden", "帳號尚無企業開通資格。")
	case errors.Is(err, organization.ErrInvalid):
		transport.WriteAPIError(w, 400, "invalid_input", "請求資料無效。")
	case errors.Is(err, organization.ErrNotFound):
		transport.WriteAPIError(w, 404, "not_found", "找不到資源。")
	case errors.Is(err, organization.ErrConflict):
		transport.WriteAPIError(w, 409, "conflict", "資源已有其他管理歸屬。")
	case err != nil:
		transport.WriteServiceError(w, 503, "unavailable", "暫時無法完成操作。", err)
	case result == nil:
		w.WriteHeader(204)
	default:
		transport.WriteJSON(w, 200, result)
	}
}

func decode(r *http.Request, w http.ResponseWriter, v any) error {
	dec := json.NewDecoder(http.MaxBytesReader(w, r.Body, 4096))
	dec.DisallowUnknownFields()
	if dec.Decode(v) != nil || dec.Decode(&struct{}{}) != io.EOF {
		return organization.ErrInvalid
	}
	return nil
}

func (h Handler) GetOrganizationContext(w http.ResponseWriter, r *http.Request) {
	h.serve(w, r, h.OrganizationStore != nil, func(ctx context.Context, user string) (any, error) {
		return h.OrganizationStore.Context(ctx, user, "", false)
	})
}

func (h Handler) SelectOrganization(w http.ResponseWriter, r *http.Request) {
	h.serve(w, r, h.OrganizationStore != nil, func(ctx context.Context, user string) (any, error) {
		var input generated.SelectOrganizationRequest
		if err := decode(r, w, &input); err != nil {
			return nil, err
		}
		return h.OrganizationStore.Context(ctx, user, input.OrganizationId, true)
	})
}

func (h Handler) CreateOrganization(w http.ResponseWriter, r *http.Request) {
	h.serve(w, r, h.OrganizationStore != nil, func(ctx context.Context, user string) (any, error) {
		var input generated.CreateOrganizationRequest
		if err := decode(r, w, &input); err != nil {
			return nil, err
		}
		return organizationResponse(h.OrganizationStore.Create(ctx, user, input.Name))
	})
}

func (h Handler) TransferOrganization(w http.ResponseWriter, r *http.Request, organizationId string) {
	h.serve(w, r, h.OrganizationStore != nil, func(ctx context.Context, user string) (any, error) {
		var input generated.TransferOrganizationRequest
		if err := decode(r, w, &input); err != nil {
			return nil, err
		}
		return organizationResponse(h.OrganizationStore.Transfer(ctx, user, organizationId, input.UserId))
	})
}

func (h Handler) SetTeamManagement(w http.ResponseWriter, r *http.Request, teamId string) {
	h.serve(w, r, h.TeamStore != nil, func(ctx context.Context, user string) (any, error) {
		var input generated.TeamManagement
		if err := decode(r, w, &input); err != nil {
			return nil, err
		}
		return nil, h.TeamStore.Manage(ctx, user, teamId, string(input.ManagementScope), input.ManagementId)
	})
}

func (h Handler) AssignEnterpriseTeam(w http.ResponseWriter, r *http.Request, organizationId, teamId string) {
	h.serve(w, r, h.TeamStore != nil, func(ctx context.Context, user string) (any, error) {
		return nil, h.TeamStore.Assign(ctx, user, organizationId, teamId, true)
	})
}

func (h Handler) RemoveEnterpriseTeam(w http.ResponseWriter, r *http.Request, organizationId, teamId string) {
	h.serve(w, r, h.TeamStore != nil, func(ctx context.Context, user string) (any, error) {
		return nil, h.TeamStore.Assign(ctx, user, organizationId, teamId, false)
	})
}

func (h Handler) ListOrganizationTeams(w http.ResponseWriter, r *http.Request, organizationId string) {
	h.serve(w, r, h.TeamStore != nil, func(ctx context.Context, user string) (any, error) {
		return teamResponse(h.TeamStore.List(ctx, user, organizationId))
	})
}

func (h Handler) CreateEnterprise(w http.ResponseWriter, r *http.Request) {
	h.serve(w, r, h.EnterpriseStore != nil, func(ctx context.Context, user string) (any, error) {
		var input generated.CreateOrganizationRequest
		if err := decode(r, w, &input); err != nil {
			return nil, err
		}
		return enterpriseLinkResponse(h.EnterpriseStore.CreateEnterprise(ctx, user, input.Name))
	})
}

func (h Handler) GetEnterprise(w http.ResponseWriter, r *http.Request, enterpriseId string) {
	h.serve(w, r, h.EnterpriseStore != nil, func(ctx context.Context, user string) (any, error) {
		return enterpriseResponse(h.EnterpriseStore.Enterprise(ctx, user, enterpriseId))
	})
}

func (h Handler) AddEnterpriseOwner(w http.ResponseWriter, r *http.Request, enterpriseId string) {
	h.serve(w, r, h.EnterpriseStore != nil, func(ctx context.Context, user string) (any, error) {
		var input generated.EnterpriseOwnerRequest
		if err := decode(r, w, &input); err != nil {
			return nil, err
		}
		return enterpriseLinkResponse(h.EnterpriseStore.UpdateEnterprise(ctx, user, enterpriseId, input.UserId, true))
	})
}

func (h Handler) AttachEnterpriseOrganization(w http.ResponseWriter, r *http.Request, enterpriseId string) {
	h.serve(w, r, h.EnterpriseStore != nil, func(ctx context.Context, user string) (any, error) {
		var input generated.SelectOrganizationRequest
		if err := decode(r, w, &input); err != nil {
			return nil, err
		}
		return enterpriseLinkResponse(h.EnterpriseStore.UpdateEnterprise(ctx, user, enterpriseId, input.OrganizationId, false))
	})
}
