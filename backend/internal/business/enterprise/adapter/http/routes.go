package enterprisehttp

import (
	"context"
	"net/http"

	"github.com/963sup/line-mini-app-bot/internal/business/enterprise/application/command"
	"github.com/963sup/line-mini-app-bot/internal/business/enterprise/application/port"
	"github.com/963sup/line-mini-app-bot/internal/business/enterprise/application/query"
	"github.com/963sup/line-mini-app-bot/internal/business/internal/orghttp"
	"github.com/963sup/line-mini-app-bot/internal/httpapi/generated"
)

type Handler struct {
	Store port.Store
	Auth  orghttp.Authenticator
}

func (h Handler) CreateEnterprise(w http.ResponseWriter, r *http.Request) {
	orghttp.Serve(w, r, h.Auth, h.Store != nil, func(ctx context.Context, user string) (any, error) {
		var input generated.CreateOrganizationRequest
		if err := orghttp.Decode(r, w, &input); err != nil {
			return nil, err
		}
		return enterpriseLinkResponse(command.Create(ctx, h.Store, user, input.Name))
	})
}

func (h Handler) GetEnterprise(w http.ResponseWriter, r *http.Request, enterpriseId string) {
	orghttp.Serve(w, r, h.Auth, h.Store != nil, func(ctx context.Context, user string) (any, error) {
		return enterpriseResponse(query.Detail(ctx, h.Store, user, enterpriseId))
	})
}

func (h Handler) AddEnterpriseOwner(w http.ResponseWriter, r *http.Request, enterpriseId string) {
	orghttp.Serve(w, r, h.Auth, h.Store != nil, func(ctx context.Context, user string) (any, error) {
		var input generated.EnterpriseOwnerRequest
		if err := orghttp.Decode(r, w, &input); err != nil {
			return nil, err
		}
		return enterpriseLinkResponse(command.AddOwner(ctx, h.Store, user, enterpriseId, input.UserId))
	})
}

func (h Handler) AttachEnterpriseOrganization(w http.ResponseWriter, r *http.Request, enterpriseId string) {
	orghttp.Serve(w, r, h.Auth, h.Store != nil, func(ctx context.Context, user string) (any, error) {
		var input generated.SelectOrganizationRequest
		if err := orghttp.Decode(r, w, &input); err != nil {
			return nil, err
		}
		return enterpriseLinkResponse(command.AttachOrganization(ctx, h.Store, user, enterpriseId, input.OrganizationId))
	})
}
