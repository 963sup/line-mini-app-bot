package port

import (
	"context"

	"github.com/963sup/line-mini-app-bot/internal/business/enterprise/application/dto"
)

type Store interface {
	CreateEnterprise(context.Context, string, string) (*dto.Link, error)
	Enterprise(context.Context, string, string) (*dto.Detail, error)
	RunOwnerChange(ctx context.Context, id, owner string, apply func(OwnerChange) (*dto.Link, error)) (*dto.Link, error)
	RunOrganizationLink(ctx context.Context, id, organization string, apply func(OrganizationLink) (*dto.Link, error)) (*dto.Link, error)
}

type EnterpriseFacts struct {
	Name                 string
	Active, ActorIsOwner bool
}

type OrganizationFacts struct {
	EnterpriseID         string
	Active, ActorIsOwner bool
}

type OwnerChange interface {
	Enterprise(actor string) (EnterpriseFacts, error)
	AddOwner() error
}

type OrganizationLink interface {
	Enterprise(actor string) (EnterpriseFacts, error)
	Organization(actor string) (OrganizationFacts, error)
	Attach() error
}
