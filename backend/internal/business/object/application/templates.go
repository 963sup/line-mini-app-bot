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
	raw, e := json.Marshal(struct {
		Name  string
		Nodes []template.Node
	}{in.Name, in.Nodes})
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

func (s *Service) InstantiateTemplate(ctx context.Context, actor, rid, id, key, parentID string, edits ...[]template.Node) (dto.TemplateInstance, error) {
	r, err := s.templateTarget(ctx, actor, rid, id, key, parentID)
	if err != nil {
		return dto.TemplateInstance{}, err
	}
	snapshot, err := s.Templates.GetTemplate(ctx, r, id)
	if err != nil {
		return dto.TemplateInstance{}, err
	}
	nodes, plan, err := PlanTemplateInstance(snapshot, edits...)
	if err != nil {
		return dto.TemplateInstance{}, err
	}
	if len(edits) != 0 {
		return s.Templates.InstantiateTemplate(ctx, r, actor, key, snapshot, parentID, plan, nodes)
	}
	return s.Templates.InstantiateTemplate(ctx, r, actor, key, snapshot, parentID, plan)
}

func (s *Service) InstantiatePersonalTemplate(ctx context.Context, actor, rid, id, key, parentID string, edits []template.Node, versions ...int64) (dto.TemplateInstance, error) {
	r, err := s.templateTarget(ctx, actor, rid, id, key, parentID)
	if err != nil {
		return dto.TemplateInstance{}, err
	}
	if s.PersonalTemplates == nil {
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
	return s.PersonalTemplates.InstantiatePersonalTemplate(ctx, r, actor, id, key, parentID, confirmed.Nodes, version)
}

func (s *Service) templateTarget(ctx context.Context, actor, rid, id, key, parentID string) (repository.Repository, error) {
	r, err := s.scope(ctx, actor, rid)
	if err != nil {
		return r, err
	}
	if s.Templates == nil {
		return r, entity.ErrUnavailable
	}
	if !repository.ValidID(id) {
		return r, entity.ErrNotFound
	}
	if !creationKey.MatchString(key) || (parentID != "" && !repository.ValidID(parentID)) {
		return r, entity.ErrInput
	}
	return r, nil
}

// PlanTemplateInstance validates the confirmed tree against its immutable source.
func PlanTemplateInstance(snapshot dto.Template, edits ...[]template.Node) ([]template.Node, []entity.Content, error) {
	if snapshot.Version < 1 || template.ValidateTree(snapshot.Nodes) != nil {
		return nil, nil, entity.ErrUnavailable
	}
	nodes := snapshot.Nodes
	if len(edits) != 0 {
		source := make(map[string]template.Node, len(snapshot.Nodes))
		for _, node := range snapshot.Nodes {
			source[node.ID] = node
		}
		retained := make([]template.Node, 0, len(edits[0]))
		for _, edit := range edits[0] {
			node, ok := source[edit.ID]
			if !ok || node.ParentID != edit.ParentID {
				return nil, nil, entity.ErrInput
			}
			node.Name, node.Quantity, node.Unit = edit.Name, edit.Quantity, edit.Unit
			retained = append(retained, node)
		}
		confirmed, err := NormalizeTemplate(dto.TemplateCreation{Name: snapshot.Name, Nodes: retained})
		if err != nil {
			return nil, nil, err
		}
		nodes = confirmed.Nodes
	}
	plan := make([]entity.Content, len(nodes))
	for i, node := range nodes {
		content, err := templateContent(node)
		if err != nil {
			return nil, nil, err
		}
		plan[i] = content
	}
	return nodes, plan, nil
}

func (s *Service) CreatePersonalTemplate(ctx context.Context, actor, key string, in dto.TemplateCreation) (dto.Template, error) {
	if actor == "" || !creationKey.MatchString(key) || (in.TemplateID == "") != (in.BaseVersion == 0) || in.BaseVersion < 0 || (in.TemplateID != "" && !repository.ValidID(in.TemplateID)) {
		return dto.Template{}, entity.ErrInput
	}
	if s == nil || s.PersonalTemplates == nil {
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
	return s.PersonalTemplates.CreatePersonalTemplate(ctx, actor, key, normalized)
}

func (s *Service) ListPersonalTemplates(ctx context.Context, actor, cursor string) (dto.TemplatePage, error) {
	if actor == "" || len(cursor) > 1024 {
		return dto.TemplatePage{}, entity.ErrInput
	}
	if s == nil || s.PersonalTemplates == nil {
		return dto.TemplatePage{}, entity.ErrUnavailable
	}
	return s.PersonalTemplates.ListPersonalTemplates(ctx, actor, cursor)
}
