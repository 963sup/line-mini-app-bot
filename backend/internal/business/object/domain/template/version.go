// Package template defines reusable engineering structures in the Object domain.
package template

import (
	"errors"
	"strings"
)

var ErrInvalidTree = errors.New("invalid object template tree")

// Node contains creation defaults for an Object. IDs are local to the version.
type Node struct {
	ID, ParentID            string
	Name, Description, Type string
	Quantity, Unit          string
}

// ValidateTree checks that nodes describe one connected, acyclic tree.
// The input order defines sibling order.
func ValidateTree(nodes []Node) error {
	byID := make(map[string]struct{}, len(nodes))
	children := make(map[string][]string, len(nodes))
	root := ""
	for _, node := range nodes {
		if node.ID == "" || strings.TrimSpace(node.Name) == "" {
			return ErrInvalidTree
		}
		if _, exists := byID[node.ID]; exists {
			return ErrInvalidTree
		}
		byID[node.ID] = struct{}{}
		if node.ParentID == "" {
			if root != "" {
				return ErrInvalidTree
			}
			root = node.ID
		} else {
			children[node.ParentID] = append(children[node.ParentID], node.ID)
		}
	}
	if root == "" {
		return ErrInvalidTree
	}
	pending := []string{root}
	visited := 0
	for len(pending) > 0 {
		id := pending[len(pending)-1]
		pending = append(pending[:len(pending)-1], children[id]...)
		visited++
	}
	// Each node has at most one parent. A cycle or a missing parent leaves
	// nodes unreachable from the root.
	if visited != len(nodes) {
		return ErrInvalidTree
	}
	return nil
}

// Version is a published snapshot scoped to a repository.
type Version struct {
	repositoryID string
	templateID   string
	number       int64
	nodes        []Node
}

func Publish(repositoryID, templateID string, number int64, draft []Node) (Version, error) {
	if repositoryID == "" || templateID == "" || number < 1 {
		return Version{}, errors.New("invalid object template version identity")
	}
	if err := ValidateTree(draft); err != nil {
		return Version{}, err
	}
	return Version{repositoryID: repositoryID, templateID: templateID, number: number, nodes: append([]Node(nil), draft...)}, nil
}

func (v Version) RepositoryID() string { return v.repositoryID }
func (v Version) TemplateID() string   { return v.templateID }
func (v Version) Number() int64        { return v.number }
func (v Version) Nodes() []Node        { return append([]Node(nil), v.nodes...) }
