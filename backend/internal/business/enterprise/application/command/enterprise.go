package command

import (
	"context"

	"github.com/963sup/line-mini-app-bot/internal/business/enterprise/application/dto"
	"github.com/963sup/line-mini-app-bot/internal/business/enterprise/application/port"
)

func Create(ctx context.Context, store port.Store, actor, name string) (*dto.Link, error) {
	return store.CreateEnterprise(ctx, actor, name)
}

func AddOwner(ctx context.Context, store port.Store, actor, id, owner string) (*dto.Link, error) {
	return store.UpdateEnterprise(ctx, actor, id, owner, true)
}

func AttachOrganization(ctx context.Context, store port.Store, actor, id, organization string) (*dto.Link, error) {
	return store.UpdateEnterprise(ctx, actor, id, organization, false)
}
