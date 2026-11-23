package application

import (
	"context"

	"github.com/963sup/line-mini-app-bot/internal/business/object/application/command"
	"github.com/963sup/line-mini-app-bot/internal/business/object/application/dto"
	"github.com/963sup/line-mini-app-bot/internal/business/object/application/port"
	"github.com/963sup/line-mini-app-bot/internal/business/object/application/query"
	"github.com/963sup/line-mini-app-bot/internal/business/object/domain/entity"
	"github.com/963sup/line-mini-app-bot/internal/business/object/domain/template"
	repository "github.com/963sup/line-mini-app-bot/internal/business/repository/domain/entity"
)

func (s *Service) templateTarget(ctx context.Context, actor, rid, id, key, parentID string) (repository.Repository, error) {
	r, err := s.scope(ctx, actor, rid)
	if err != nil {
		return r, err
	}
	if !repository.ValidID(id) {
		return r, entity.ErrNotFound
	}
	if !command.ValidCreationKey(key) || (parentID != "" && !repository.ValidID(parentID)) {
		return r, entity.ErrInput
	}
	return r, nil
}

func (s *Service) CreateTemplate(ctx context.Context, actor, rid, key string, in dto.TemplateCreation) (dto.Template, error) {
	r, e := s.scope(ctx, actor, rid)
	if e != nil {
		return dto.Template{}, e
	}
	return command.CreateTemplate(ctx, s.Templates, r, actor, key, in)
}

func (s *Service) ListTemplates(ctx context.Context, actor, rid, cursor string) (dto.TemplatePage, error) {
	r, e := s.scope(ctx, actor, rid)
	if e != nil {
		return dto.TemplatePage{}, e
	}
	return query.ListTemplates(ctx, s.Templates, r, actor, cursor)
}

func (s *Service) GetTemplate(ctx context.Context, actor, rid, id string) (dto.Template, error) {
	r, e := s.scope(ctx, actor, rid)
	if e != nil {
		return dto.Template{}, e
	}
	return query.GetTemplate(ctx, s.Templates, r, id)
}

func (s *Service) InstantiateTemplate(ctx context.Context, actor, rid, id, key, parentID string, edits ...[]template.Node) (dto.TemplateInstance, error) {
	r, err := s.templateTarget(ctx, actor, rid, id, key, parentID)
	if err != nil {
		return dto.TemplateInstance{}, err
	}
	return command.InstantiateTemplate(ctx, s.Templates, r, actor, id, key, parentID, edits...)
}

func (s *Service) InstantiatePersonalTemplate(ctx context.Context, actor, rid, id, key, parentID string, edits []template.Node, versions ...int64) (dto.TemplateInstance, error) {
	r, err := s.templateTarget(ctx, actor, rid, id, key, parentID)
	if err != nil {
		return dto.TemplateInstance{}, err
	}
	return command.InstantiatePersonalTemplate(ctx, s.PersonalTemplates, r, actor, id, key, parentID, edits, versions...)
}

func (s *Service) CreatePersonalTemplate(ctx context.Context, actor, key string, in dto.TemplateCreation) (dto.Template, error) {
	var store port.PersonalTemplateStore
	if s != nil {
		store = s.PersonalTemplates
	}
	return command.CreatePersonalTemplate(ctx, store, actor, key, in)
}

func (s *Service) ListPersonalTemplates(ctx context.Context, actor, cursor string) (dto.TemplatePage, error) {
	var store port.PersonalTemplateStore
	if s != nil {
		store = s.PersonalTemplates
	}
	return query.ListPersonalTemplates(ctx, store, actor, cursor)
}
