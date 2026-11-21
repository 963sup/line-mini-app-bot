package application

import (
	"context"
	"encoding/json"
	"regexp"
	"strings"
	"unicode/utf8"

	"github.com/963sup/line-mini-app-bot/internal/business/object/application/dto"
	"github.com/963sup/line-mini-app-bot/internal/business/object/domain/entity"
	"github.com/963sup/line-mini-app-bot/internal/business/object/domain/template"
	repository "github.com/963sup/line-mini-app-bot/internal/business/repository/domain/entity"
)

const MaxTemplateNodes = 100

var templateNodeID = regexp.MustCompile(`^[A-Za-z0-9_-]{1,64}$`)

// NormalizeTemplate validates defaults through the same Object rules used at creation.
// The template domain itself stays independent of the Object entity.
func NormalizeTemplate(in dto.TemplateCreation) (dto.TemplateCreation, error) {
	in.Name = strings.TrimSpace(in.Name)
	if !utf8.ValidString(in.Name) || in.Name == "" || utf8.RuneCountInString(in.Name) > 120 || len(in.Nodes) == 0 || len(in.Nodes) > MaxTemplateNodes {
		return dto.TemplateCreation{}, entity.ErrInput
	}
	in.Nodes = append([]template.Node(nil), in.Nodes...)
	for i, n := range in.Nodes {
		if !templateNodeID.MatchString(n.ID) || (n.ParentID != "" && !templateNodeID.MatchString(n.ParentID)) {
			return dto.TemplateCreation{}, entity.ErrInput
		}
		c, e := templateContent(n)
		if e != nil {
			return dto.TemplateCreation{}, e
		}
		in.Nodes[i] = template.Node{ID: n.ID, ParentID: n.ParentID, Name: c.Name, Description: c.Description, Type: c.Type, Quantity: c.Quantity, Unit: c.Unit}
	}
	if template.ValidateTree(in.Nodes) != nil {
		return dto.TemplateCreation{}, entity.ErrInput
	}
	raw, e := json.Marshal(in)
	if e != nil || len(raw) > 48000 {
		return dto.TemplateCreation{}, entity.ErrInput
	}
	return in, nil
}

func templateContent(n template.Node) (entity.Content, error) {
	c := entity.Content{Name: n.Name, Description: n.Description, Type: n.Type, Quantity: n.Quantity, Unit: n.Unit}
	if strings.TrimSpace(c.Type) == entity.TypeWork && strings.TrimSpace(c.Quantity) != "" {
		c.CompletedQuantity = "0"
	}
	return entity.Normalize(c)
}

func (s *Service) CreateTemplate(ctx context.Context, actor, rid, key string, in dto.TemplateCreation) (dto.Template, error) {
	r, e := s.scope(ctx, actor, rid)
	if e != nil {
		return dto.Template{}, e
	}
	if s.Templates == nil {
		return dto.Template{}, entity.ErrUnavailable
	}
	if !creationKey.MatchString(key) {
		return dto.Template{}, entity.ErrInput
	}
	in, e = NormalizeTemplate(in)
	if e != nil {
		return dto.Template{}, e
	}
	return s.Templates.CreateTemplate(ctx, r, actor, key, in)
}

func (s *Service) ListTemplates(ctx context.Context, actor, rid, cursor string) (dto.TemplatePage, error) {
	r, e := s.scope(ctx, actor, rid)
	if e != nil {
		return dto.TemplatePage{}, e
	}
	if s.Templates == nil {
		return dto.TemplatePage{}, entity.ErrUnavailable
	}
	if len(cursor) > 1024 {
		return dto.TemplatePage{}, entity.ErrInput
	}
	return s.Templates.ListTemplates(ctx, r, actor, cursor)
}

func (s *Service) GetTemplate(ctx context.Context, actor, rid, id string) (dto.Template, error) {
	r, e := s.scope(ctx, actor, rid)
	if e != nil {
		return dto.Template{}, e
	}
	if s.Templates == nil {
		return dto.Template{}, entity.ErrUnavailable
	}
	if !repository.ValidID(id) {
		return dto.Template{}, entity.ErrNotFound
	}
	return s.Templates.GetTemplate(ctx, r, id)
}

func (s *Service) InstantiateTemplate(ctx context.Context, actor, rid, id, key, parentID string) (dto.TemplateInstance, error) {
	r, e := s.scope(ctx, actor, rid)
	if e != nil {
		return dto.TemplateInstance{}, e
	}
	if s.Templates == nil {
		return dto.TemplateInstance{}, entity.ErrUnavailable
	}
	if !repository.ValidID(id) {
		return dto.TemplateInstance{}, entity.ErrNotFound
	}
	if !creationKey.MatchString(key) || (parentID != "" && !repository.ValidID(parentID)) {
		return dto.TemplateInstance{}, entity.ErrInput
	}
	snapshot, e := s.Templates.GetTemplate(ctx, r, id)
	if e != nil {
		return dto.TemplateInstance{}, e
	}
	if _, e = template.Publish(rid, id, snapshot.Version, snapshot.Nodes); e != nil {
		return dto.TemplateInstance{}, entity.ErrUnavailable
	}
	plan := make([]entity.Content, len(snapshot.Nodes))
	for i, n := range snapshot.Nodes {
		plan[i], e = templateContent(n)
		if e != nil {
			return dto.TemplateInstance{}, entity.ErrUnavailable
		}
	}
	return s.Templates.InstantiateTemplate(ctx, r, actor, key, snapshot, parentID, plan)
}
