package query

import (
	"context"

	"github.com/963sup/line-mini-app-bot/internal/business/object/application/dto"
	"github.com/963sup/line-mini-app-bot/internal/business/object/application/port"
	"github.com/963sup/line-mini-app-bot/internal/business/object/domain/entity"
	repository "github.com/963sup/line-mini-app-bot/internal/business/repository/domain/entity"
)

func ListTemplates(ctx context.Context, store port.TemplateStore, r repository.Repository, actor string, cursor string) (dto.TemplatePage, error) {
	if store == nil {
		return dto.TemplatePage{}, entity.ErrUnavailable
	}
	if len(cursor) > 1024 {
		return dto.TemplatePage{}, entity.ErrInput
	}
	return store.ListTemplates(ctx, r, actor, cursor)
}

func GetTemplate(ctx context.Context, store port.TemplateStore, r repository.Repository, id string) (dto.Template, error) {
	if store == nil {
		return dto.Template{}, entity.ErrUnavailable
	}
	if !repository.ValidID(id) {
		return dto.Template{}, entity.ErrNotFound
	}
	return store.GetTemplate(ctx, r, id)
}

func ListPersonalTemplates(ctx context.Context, store port.PersonalTemplateStore, actor string, cursor string) (dto.TemplatePage, error) {
	if actor == "" || len(cursor) > 1024 {
		return dto.TemplatePage{}, entity.ErrInput
	}
	if store == nil {
		return dto.TemplatePage{}, entity.ErrUnavailable
	}
	return store.ListPersonalTemplates(ctx, actor, cursor)
}
