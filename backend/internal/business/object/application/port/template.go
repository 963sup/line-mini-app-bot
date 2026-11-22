package port

import (
	"context"

	"github.com/963sup/line-mini-app-bot/internal/business/object/application/dto"
	"github.com/963sup/line-mini-app-bot/internal/business/object/domain/entity"
	"github.com/963sup/line-mini-app-bot/internal/business/object/domain/template"
	repository "github.com/963sup/line-mini-app-bot/internal/business/repository/domain/entity"
)

// TemplateStore owns atomic snapshot creation and whole-tree commits with receipts.
// Implementations recheck the repository container and target parent before writing.
type TemplateStore interface {
	CreateTemplate(context.Context, repository.Repository, string, string, dto.TemplateCreation) (dto.Template, error)
	ListTemplates(context.Context, repository.Repository, string, string) (dto.TemplatePage, error)
	GetTemplate(context.Context, repository.Repository, string) (dto.Template, error)
	InstantiateTemplate(context.Context, repository.Repository, string, string, dto.Template, string, []entity.Content, ...[]template.Node) (dto.TemplateInstance, error)
}

// PersonalTemplateStore scopes snapshots to the authenticated actor.
type PersonalTemplateStore interface {
	CreatePersonalTemplate(context.Context, string, string, dto.TemplateCreation) (dto.Template, error)
	ListPersonalTemplates(context.Context, string, string) (dto.TemplatePage, error)
	InstantiatePersonalTemplate(context.Context, repository.Repository, string, string, string, string, []template.Node) (dto.TemplateInstance, error)
}
