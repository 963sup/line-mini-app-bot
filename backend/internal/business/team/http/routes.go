package teamhttp

import (
	"context"
	"net/http"

	"github.com/963sup/line-mini-app-bot/internal/business/internal/adapter/http"
	"github.com/963sup/line-mini-app-bot/internal/business/team"
	"github.com/963sup/line-mini-app-bot/internal/httpapi/generated"
)

type Store interface {
	Manage(ctx context.Context, user string, teamId, scope, managementId string) error
	Assign(ctx context.Context, user string, id, teamId string, add bool) error
	List(ctx context.Context, user string, organizationId string) (*team.Listing, error)
}

type Handler struct {
	Store Store
	Auth  orghttp.Authenticator
}

func (h Handler) SetTeamManagement(w http.ResponseWriter, r *http.Request, teamId string) {
	orghttp.Serve(w, r, h.Auth, h.Store != nil, func(ctx context.Context, user string) (any, error) {
		var input generated.TeamManagement
		if err := orghttp.Decode(r, w, &input); err != nil {
			return nil, err
		}
		return nil, h.Store.Manage(ctx, user, teamId, string(input.ManagementScope), input.ManagementId)
	})
}

func (h Handler) AssignEnterpriseTeam(w http.ResponseWriter, r *http.Request, organizationId, teamId string) {
	orghttp.Serve(w, r, h.Auth, h.Store != nil, func(ctx context.Context, user string) (any, error) {
		return nil, h.Store.Assign(ctx, user, organizationId, teamId, true)
	})
}

func (h Handler) RemoveEnterpriseTeam(w http.ResponseWriter, r *http.Request, organizationId, teamId string) {
	orghttp.Serve(w, r, h.Auth, h.Store != nil, func(ctx context.Context, user string) (any, error) {
		return nil, h.Store.Assign(ctx, user, organizationId, teamId, false)
	})
}

func (h Handler) ListOrganizationTeams(w http.ResponseWriter, r *http.Request, organizationId string) {
	orghttp.Serve(w, r, h.Auth, h.Store != nil, func(ctx context.Context, user string) (any, error) {
		return teamResponse(h.Store.List(ctx, user, organizationId))
	})
}
