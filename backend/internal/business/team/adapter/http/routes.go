package teamhttp

import (
	"context"
	"errors"
	"net/http"

	"github.com/963sup/line-mini-app-bot/internal/business/internal/adapter/http"
	organizationmodel "github.com/963sup/line-mini-app-bot/internal/business/organization/domain/entity"
	teamapp "github.com/963sup/line-mini-app-bot/internal/business/team/application"
	teamApplicationDto "github.com/963sup/line-mini-app-bot/internal/business/team/application/dto"
	"github.com/963sup/line-mini-app-bot/internal/transport/http/api/generated"
)

type Store interface {
	Activate(ctx context.Context, actor, teamID, scope, managementID, maintainer string) error
	Manage(ctx context.Context, user string, teamId, scope, managementId string) error
	teamapp.AssignmentStore
	List(ctx context.Context, user string, organizationId string) (*teamApplicationDto.Listing, error)
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
		if input.InitialMaintainerUserId != nil {
			return nil, h.Store.Activate(ctx, user, teamId, string(input.ManagementScope), input.ManagementId, *input.InitialMaintainerUserId)
		}
		return nil, h.Store.Manage(ctx, user, teamId, string(input.ManagementScope), input.ManagementId)
	})
}

func (h Handler) AssignEnterpriseTeam(w http.ResponseWriter, r *http.Request, organizationId, teamId string) {
	h.changeAssignment(w, r, organizationId, teamId, true)
}

func (h Handler) RemoveEnterpriseTeam(w http.ResponseWriter, r *http.Request, organizationId, teamId string) {
	h.changeAssignment(w, r, organizationId, teamId, false)
}

func (h Handler) changeAssignment(w http.ResponseWriter, r *http.Request, organizationID, teamID string, assigned bool) {
	orghttp.Serve(w, r, h.Auth, h.Store != nil, func(ctx context.Context, user string) (any, error) {
		err := teamapp.AssignEnterpriseTeam(ctx, h.Store, user, organizationID, teamID, assigned)
		if errors.Is(err, teamapp.ErrAssignmentNotFound) {
			return nil, organizationmodel.ErrNotFound
		}
		return nil, err
	})
}

func (h Handler) ListOrganizationTeams(w http.ResponseWriter, r *http.Request, organizationId string) {
	orghttp.Serve(w, r, h.Auth, h.Store != nil, func(ctx context.Context, user string) (any, error) {
		return teamResponse(h.Store.List(ctx, user, organizationId))
	})
}
