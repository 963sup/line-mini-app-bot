package objecthttp

import (
	"errors"
	"net/http"

	"github.com/963sup/line-mini-app-bot/internal/business/object/domain/entity"

	"github.com/963sup/line-mini-app-bot/internal/business/object/application/dto"
	"github.com/963sup/line-mini-app-bot/internal/business/object/domain/template"
	"github.com/963sup/line-mini-app-bot/internal/identity"
	transport "github.com/963sup/line-mini-app-bot/internal/transport/http"
	"github.com/963sup/line-mini-app-bot/internal/transport/http/api/generated"
)

func templateView(t dto.Template) generated.ObjectTemplate {
	nodes := make([]generated.ObjectTemplateNode, len(t.Nodes))
	for i, n := range t.Nodes {
		nodes[i] = generated.ObjectTemplateNode{Id: n.ID, Name: n.Name, ParentId: optional(n.ParentID), Description: optional(n.Description), Type: optional(n.Type), Quantity: optional(n.Quantity), Unit: optional(n.Unit)}
	}
	return generated.ObjectTemplate{Id: t.ID, RepositoryId: t.RepositoryID, Name: t.Name, CreatorId: t.CreatorID, Version: 1, Nodes: nodes, CreatedAt: t.CreatedAt}
}

func optional(s string) *string {
	if s == "" {
		return nil
	}
	return &s
}

func (s *objectServer) CreateObjectTemplate(w http.ResponseWriter, r *http.Request, rid string, p generated.CreateObjectTemplateParams) {
	var in generated.ObjectTemplateCreation
	if !transport.ReadJSON(w, r, &in, "範本資料無效。") {
		return
	}
	nodes := make([]template.Node, len(in.Nodes))
	for i, n := range in.Nodes {
		nodes[i] = template.Node{ID: n.Id, ParentID: value(n.ParentId), Name: n.Name, Description: value(n.Description), Type: value(n.Type), Quantity: value(n.Quantity), Unit: value(n.Unit)}
	}
	actor := identity.UserID(r.Context())
	input := dto.TemplateCreation{Name: in.Name, Nodes: nodes}
	var t dto.Template
	var e error
	if p.Scope != nil && *p.Scope == "personal" {
		t, e = s.service.CreatePersonalTemplate(r.Context(), actor, p.IdempotencyKey, input)
	} else {
		t, e = s.service.CreateTemplate(r.Context(), actor, rid, p.IdempotencyKey, input)
	}
	if !writeTemplateError(w, e) {
		transport.WriteJSON(w, 200, templateView(t))
	}
}

func (s *objectServer) ListObjectTemplates(w http.ResponseWriter, r *http.Request, rid string, p generated.ListObjectTemplatesParams) {
	actor := identity.UserID(r.Context())
	var page dto.TemplatePage
	var e error
	if p.Scope != nil && *p.Scope == "personal" {
		page, e = s.service.ListPersonalTemplates(r.Context(), actor, value(p.Cursor))
	} else {
		page, e = s.service.ListTemplates(r.Context(), actor, rid, value(p.Cursor))
	}
	if writeTemplateError(w, e) {
		return
	}
	items := make([]generated.ObjectTemplate, len(page.Items))
	for i, t := range page.Items {
		items[i] = templateView(t)
	}
	transport.WriteJSON(w, 200, generated.ObjectTemplatePage{Items: items, NextCursor: page.NextCursor})
}

func (s *objectServer) GetObjectTemplate(w http.ResponseWriter, r *http.Request, rid, id string) {
	t, e := s.service.GetTemplate(r.Context(), identity.UserID(r.Context()), rid, id)
	if !writeTemplateError(w, e) {
		transport.WriteJSON(w, 200, templateView(t))
	}
}

func (s *objectServer) InstantiateObjectTemplate(w http.ResponseWriter, r *http.Request, rid, id string, p generated.InstantiateObjectTemplateParams) {
	var in generated.ObjectTemplateInstantiation
	if !transport.ReadJSON(w, r, &in, "套用範本資料無效。") {
		return
	}
	var edits [][]template.Node
	if in.Nodes != nil {
		nodes := make([]template.Node, 0, len(*in.Nodes))
		for _, node := range *in.Nodes {
			nodes = append(nodes, template.Node{ID: node.Id, ParentID: value(node.ParentId), Name: node.Name, Quantity: value(node.Quantity), Unit: value(node.Unit)})
		}
		edits = append(edits, nodes)
	}
	actor := identity.UserID(r.Context())
	var result dto.TemplateInstance
	var e error
	if p.Scope != nil && *p.Scope == "personal" {
		if len(edits) == 0 {
			writeTemplateError(w, entity.ErrInput)
			return
		}
		result, e = s.service.InstantiatePersonalTemplate(r.Context(), actor, rid, id, p.IdempotencyKey, value(in.ParentId), edits[0])
	} else {
		result, e = s.service.InstantiateTemplate(r.Context(), actor, rid, id, p.IdempotencyKey, value(in.ParentId), edits...)
	}
	if !writeTemplateError(w, e) {
		nodes := make([]generated.ObjectTemplateInstanceNode, len(result.Nodes))
		for i, n := range result.Nodes {
			nodes[i] = generated.ObjectTemplateInstanceNode{ObjectId: n.ObjectID, ParentId: n.ParentID, Name: n.Name}
		}
		transport.WriteJSON(w, 200, generated.ObjectTemplateInstance{Nodes: nodes, RootObjectId: result.RootObjectID, ObjectIds: result.ObjectIDs, TemplateId: result.TemplateID, TemplateVersion: result.TemplateVersion})
	}
}

func writeTemplateError(w http.ResponseWriter, e error) bool {
	if errors.Is(e, entity.ErrInput) {
		transport.WriteAPIError(w, 400, "invalid_input", "請確認範本名稱、單一根節點與父子關係；最多100個節點、48,000 bytes，數量與單位須符合物件規則。")
		return true
	}
	return writeError(w, e)
}

func (s *objectServer) ListPersonalObjectTemplates(w http.ResponseWriter, r *http.Request, p generated.ListPersonalObjectTemplatesParams) {
	scope := generated.ListObjectTemplatesParamsScopePersonal
	s.ListObjectTemplates(w, r, "", generated.ListObjectTemplatesParams{Scope: &scope, Cursor: p.Cursor})
}

func (s *objectServer) CreatePersonalObjectTemplate(w http.ResponseWriter, r *http.Request, p generated.CreatePersonalObjectTemplateParams) {
	scope := generated.CreateObjectTemplateParamsScopePersonal
	s.CreateObjectTemplate(w, r, "", generated.CreateObjectTemplateParams{Scope: &scope, IdempotencyKey: p.IdempotencyKey})
}
