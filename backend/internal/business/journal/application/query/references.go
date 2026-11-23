package query

import (
	"context"

	groupdto "github.com/963sup/line-mini-app-bot/internal/linegroup/application/dto"
	groupentity "github.com/963sup/line-mini-app-bot/internal/linegroup/domain/entity"
)

// GroupCatalog supplies the current product membership for both reads and listings.
type GroupCatalog interface {
	Get(context.Context, string, string) (groupentity.Group, error)
	List(context.Context, string, string) (groupdto.Page, error)
}

// ReferenceCatalog provides access checks for journal references.
type ReferenceCatalog struct {
	Groups          GroupCatalog
	RepositoryOwner func(ctx context.Context, actor, id string) (string, error)
	ProjectOwner    func(ctx context.Context, actor, group, id string) (string, error)
}
