package query

import (
	"context"

	project "github.com/963sup/line-mini-app-bot/internal/business/project/application"
	repository "github.com/963sup/line-mini-app-bot/internal/business/repository/application"
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
	Groups       GroupCatalog
	Repositories *repository.Service
	Projects     *project.Service
}
