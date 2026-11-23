package command

import (
	"context"
	"strings"

	"github.com/963sup/line-mini-app-bot/internal/business/object/application/dto"
	"github.com/963sup/line-mini-app-bot/internal/business/object/application/port"
	"github.com/963sup/line-mini-app-bot/internal/business/object/domain/entity"
	"github.com/963sup/line-mini-app-bot/internal/business/object/domain/service"
	"github.com/963sup/line-mini-app-bot/internal/business/object/domain/template"
	repository "github.com/963sup/line-mini-app-bot/internal/business/repository/domain/entity"
)

func NormalizeTemplate(in dto.TemplateCreation) (dto.TemplateCreation, error) {
	definition, err := service.NormalizeTemplate(service.Definition{Name: in.Name, Nodes: in.Nodes})
	if err != nil {
		return dto.TemplateCreation{}, err
	}
	in.Name, in.Nodes = definition.Name, definition.Nodes
	return in, nil
}

func CreateTemplate(ctx context.Context, store port.TemplateStore, r repository.Repository, actor string, key string, in dto.TemplateCreation) (dto.Template, error) {
	if store == nil {
		return dto.Template{}, entity.ErrUnavailable
	}
	if !creationKey.MatchString(key) {
		return dto.Template{}, entity.ErrInput
	}
	in, e := NormalizeTemplate(in)
	if e != nil {
		return dto.Template{}, e
	}
	return store.CreateTemplate(ctx, r, actor, key, in)
}

func InstantiateTemplate(ctx context.Context, store port.TemplateStore, r repository.Repository, actor string, id string, key string, parentID string, edits ...[]template.Node) (dto.TemplateInstance, error) {
	if store == nil {
		return dto.TemplateInstance{}, entity.ErrUnavailable
	}
	snapshot, err := store.GetTemplate(ctx, r, id)
	if err != nil {
		return dto.TemplateInstance{}, err
	}
	nodes, plan, err := service.PlanTemplateInstance(snapshot.Version, service.Definition{Name: snapshot.Name, Nodes: snapshot.Nodes}, edits...)
	if err != nil {
		return dto.TemplateInstance{}, err
	}
	if len(edits) != 0 {
		return store.InstantiateTemplate(ctx, r, actor, key, snapshot, parentID, plan, nodes)
	}
	return store.InstantiateTemplate(ctx, r, actor, key, snapshot, parentID, plan)
}

func InstantiatePersonalTemplate(ctx context.Context, store port.PersonalTemplateStore, r repository.Repository, actor string, id string, key string, parentID string, edits []template.Node, versions ...int64) (dto.TemplateInstance, error) {
	if store == nil {
		return dto.TemplateInstance{}, entity.ErrUnavailable
	}
	// Normalize only editable values before computing the retry fingerprint.
	retained := make([]template.Node, len(edits))
	for i, n := range edits {
		retained[i] = template.Node{ID: n.ID, ParentID: n.ParentID, Name: n.Name, Quantity: n.Quantity, Unit: n.Unit}
	}
	confirmed, err := NormalizeTemplate(dto.TemplateCreation{Name: "instance", Nodes: retained})
	if err != nil {
		return dto.TemplateInstance{}, err
	}
	version := int64(1)
	if len(versions) > 0 {
		version = versions[0]
	}
	if version < 1 {
		return dto.TemplateInstance{}, entity.ErrInput
	}
	return store.InstantiatePersonalTemplate(ctx, r, actor, id, key, parentID, confirmed.Nodes, version)
}

func CreatePersonalTemplate(ctx context.Context, store port.PersonalTemplateStore, actor string, key string, in dto.TemplateCreation) (dto.Template, error) {
	if actor == "" || !creationKey.MatchString(key) || (in.TemplateID == "") != (in.BaseVersion == 0) || in.BaseVersion < 0 || (in.TemplateID != "" && !repository.ValidID(in.TemplateID)) {
		return dto.Template{}, entity.ErrInput
	}
	if store == nil {
		return dto.Template{}, entity.ErrUnavailable
	}
	for _, n := range in.Nodes {
		if strings.TrimSpace(n.Quantity) != "" {
			return dto.Template{}, entity.ErrInput
		}
	}
	normalized, err := NormalizeTemplate(in)
	if err != nil {
		return dto.Template{}, err
	}
	return store.CreatePersonalTemplate(ctx, actor, key, normalized)
}
