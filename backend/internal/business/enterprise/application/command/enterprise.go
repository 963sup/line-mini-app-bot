package command

import (
	"context"

	enterprise "github.com/963sup/line-mini-app-bot/internal/business/enterprise/domain"

	"github.com/963sup/line-mini-app-bot/internal/business/enterprise/application/dto"
	"github.com/963sup/line-mini-app-bot/internal/business/enterprise/application/port"
)

func Create(ctx context.Context, store port.Store, actor, name string) (*dto.Link, error) {
	return store.CreateEnterprise(ctx, actor, name)
}

func AddOwner(ctx context.Context, store port.Store, actor, id, owner string) (*dto.Link, error) {
	return store.RunOwnerChange(ctx, id, owner, func(work port.OwnerChange) (*dto.Link, error) {
		ent, err := work.Enterprise(actor)
		if err != nil {
			return nil, err
		}
		if !ent.Active || !ent.ActorIsOwner {
			return nil, enterprise.ErrNotFound
		}
		return &dto.Link{ID: id, Name: ent.Name}, work.AddOwner()
	})
}

func AttachOrganization(ctx context.Context, store port.Store, actor, id, organizationID string) (*dto.Link, error) {
	return store.RunOrganizationLink(ctx, id, organizationID, func(work port.OrganizationLink) (*dto.Link, error) {
		ent, err := work.Enterprise(actor)
		if err != nil {
			return nil, err
		}
		if !ent.Active || !ent.ActorIsOwner {
			return nil, enterprise.ErrNotFound
		}
		org, err := work.Organization(actor)
		if err != nil {
			return nil, err
		}
		if !org.Active || !org.ActorIsOwner {
			return nil, enterprise.ErrNotFound
		}
		if org.EnterpriseID != "" && org.EnterpriseID != id {
			return nil, enterprise.ErrConflict
		}
		return &dto.Link{ID: id, Name: ent.Name}, work.Attach()
	})
}
