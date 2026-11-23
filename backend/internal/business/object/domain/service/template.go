// Package service defines rules that combine templates and Object creation.
package service

import (
	"encoding/json"
	"regexp"
	"strings"
	"unicode/utf8"

	"github.com/963sup/line-mini-app-bot/internal/business/object/domain/entity"
	"github.com/963sup/line-mini-app-bot/internal/business/object/domain/template"
)

// Definition contains the creation defaults for one named tree.
type Definition struct {
	Name  string
	Nodes []template.Node
}

const MaxTemplateNodes = 100

var templateNodeID = regexp.MustCompile(`^[A-Za-z0-9_-]{1,64}$`)

// NormalizeTemplate validates defaults through the same Object rules used at creation.
// The template domain itself stays independent of the Object entity.
func NormalizeTemplate(in Definition) (Definition, error) {
	in.Name = strings.TrimSpace(in.Name)
	if !utf8.ValidString(in.Name) || in.Name == "" || utf8.RuneCountInString(in.Name) > 120 || len(in.Nodes) == 0 || len(in.Nodes) > MaxTemplateNodes {
		return Definition{}, entity.ErrInput
	}
	in.Nodes = append([]template.Node(nil), in.Nodes...)
	for i, n := range in.Nodes {
		if !templateNodeID.MatchString(n.ID) || (n.ParentID != "" && !templateNodeID.MatchString(n.ParentID)) {
			return Definition{}, entity.ErrInput
		}
		c, e := templateContent(n)
		if e != nil {
			return Definition{}, e
		}
		in.Nodes[i] = template.Node{ID: n.ID, ParentID: n.ParentID, Name: c.Name, Description: c.Description, Type: c.Type, Quantity: c.Quantity, Unit: c.Unit}
	}
	if template.ValidateTree(in.Nodes) != nil {
		return Definition{}, entity.ErrInput
	}
	raw, e := json.Marshal(in)
	if e != nil || len(raw) > 48000 {
		return Definition{}, entity.ErrInput
	}
	return in, nil
}

func templateContent(n template.Node) (entity.Content, error) {
	c := entity.Content{Name: n.Name, Description: n.Description, Type: n.Type, Quantity: n.Quantity, Unit: n.Unit}
	return entity.NormalizeCreation(c)
}

// PlanTemplateInstance validates the confirmed tree against its immutable source.
func PlanTemplateInstance(version int64, sourceTree Definition, edits ...[]template.Node) ([]template.Node, []entity.Content, error) {
	if version < 1 || template.ValidateTree(sourceTree.Nodes) != nil {
		return nil, nil, entity.ErrUnavailable
	}
	nodes := sourceTree.Nodes
	if len(edits) != 0 {
		source := make(map[string]template.Node, len(sourceTree.Nodes))
		for _, node := range sourceTree.Nodes {
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
		confirmed, err := NormalizeTemplate(Definition{Name: sourceTree.Name, Nodes: retained})
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
