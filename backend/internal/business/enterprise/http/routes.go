package enterprisehttp

import (
	"context"
	"github.com/963sup/line-mini-app-bot/internal/business/enterprise"
	"github.com/963sup/line-mini-app-bot/internal/business/internal/orghttp"
	"github.com/963sup/line-mini-app-bot/internal/httpapi/generated"
	"net/http"
)

type Store interface {
	CreateEnterprise(context.Context, string, string) (*enterprise.Link, error)
	Enterprise(context.Context, string, string) (*enterprise.Detail, error)
	UpdateEnterprise(context.Context, string, string, string, bool) (*enterprise.Link, error)
}

type Handler struct {
	Store Store
	Auth  orghttp.Authenticator
}

func (h Handler) CreateEnterprise(w http.ResponseWriter, r *http.Request) {
	orghttp.Serve(w, r, h.Auth, h.Store != nil, func(ctx context.Context, user string) (any, error) {
		var input generated.CreateOrganizationRequest
		if err := orghttp.Decode(r, w, &input); err != nil {
			return nil, err
		}
		return enterpriseLinkResponse(h.Store.CreateEnterprise(ctx, user, input.Name))
	})
}

func (h Handler) GetEnterprise(w http.ResponseWriter, r *http.Request, enterpriseId string) {
	orghttp.Serve(w, r, h.Auth, h.Store != nil, func(ctx context.Context, user string) (any, error) {
		return enterpriseResponse(h.Store.Enterprise(ctx, user, enterpriseId))
	})
}

func (h Handler) AddEnterpriseOwner(w http.ResponseWriter, r *http.Request, enterpriseId string) {
	orghttp.Serve(w, r, h.Auth, h.Store != nil, func(ctx context.Context, user string) (any, error) {
		var input generated.EnterpriseOwnerRequest
		if err := orghttp.Decode(r, w, &input); err != nil {
			return nil, err
		}
		return enterpriseLinkResponse(h.Store.UpdateEnterprise(ctx, user, enterpriseId, input.UserId, true))
	})
}

func (h Handler) AttachEnterpriseOrganization(w http.ResponseWriter, r *http.Request, enterpriseId string) {
	orghttp.Serve(w, r, h.Auth, h.Store != nil, func(ctx context.Context, user string) (any, error) {
		var input generated.SelectOrganizationRequest
		if err := orghttp.Decode(r, w, &input); err != nil {
			return nil, err
		}
		return enterpriseLinkResponse(h.Store.UpdateEnterprise(ctx, user, enterpriseId, input.OrganizationId, false))
	})
}
