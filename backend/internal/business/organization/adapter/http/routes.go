package organizationhttp

import (
	organizationApplicationDto "github.com/963sup/line-mini-app-bot/internal/business/organization/application/dto"
	organizationDomainEntity "github.com/963sup/line-mini-app-bot/internal/business/organization/domain/entity"
)

import (
	"context"
	"net/http"

	"github.com/963sup/line-mini-app-bot/internal/business/internal/adapter/http"

	"github.com/963sup/line-mini-app-bot/internal/transport/http/api/generated"
)

type Store interface {
	Context(ctx context.Context, user string, preference string, write bool) (*organizationApplicationDto.Context, error)
	Create(ctx context.Context, user string, name string) (*organizationDomainEntity.Organization, error)
	Transfer(ctx context.Context, user string, organizationId, target string) (*organizationDomainEntity.Organization, error)
}

type Handler struct {
	Store Store
	Auth  orghttp.Authenticator
}

func (h Handler) GetOrganizationContext(w http.ResponseWriter, r *http.Request) {
	orghttp.Serve(w, r, h.Auth, h.Store != nil, func(ctx context.Context, user string) (any, error) {
		return h.Store.Context(ctx, user, "", false)
	})
}

func (h Handler) SelectOrganization(w http.ResponseWriter, r *http.Request) {
	orghttp.Serve(w, r, h.Auth, h.Store != nil, func(ctx context.Context, user string) (any, error) {
		var input generated.SelectOrganizationRequest
		if err := orghttp.Decode(r, w, &input); err != nil {
			return nil, err
		}
		return h.Store.Context(ctx, user, input.OrganizationId, true)
	})
}

func (h Handler) CreateOrganization(w http.ResponseWriter, r *http.Request) {
	orghttp.Serve(w, r, h.Auth, h.Store != nil, func(ctx context.Context, user string) (any, error) {
		var input generated.CreateOrganizationRequest
		if err := orghttp.Decode(r, w, &input); err != nil {
			return nil, err
		}
		return organizationResponse(h.Store.Create(ctx, user, input.Name))
	})
}

func (h Handler) TransferOrganization(w http.ResponseWriter, r *http.Request, organizationId string) {
	orghttp.Serve(w, r, h.Auth, h.Store != nil, func(ctx context.Context, user string) (any, error) {
		var input generated.TransferOrganizationRequest
		if err := orghttp.Decode(r, w, &input); err != nil {
			return nil, err
		}
		return organizationResponse(h.Store.Transfer(ctx, user, organizationId, input.UserId))
	})
}
