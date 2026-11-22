package dto

import (
	"time"

	"github.com/963sup/line-mini-app-bot/internal/business/object/domain/template"
)

// Template is an published snapshot of one template version.
type Template struct {
	ID, RepositoryID, Name, CreatorID string
	Version                           int64
	Nodes                             []template.Node
	CreatedAt                         time.Time
}

type TemplateCreation struct {
	TemplateID  string
	BaseVersion int64
	Name        string
	Nodes       []template.Node
}

type TemplatePage struct {
	Items      []Template
	NextCursor string
}

type TemplateInstanceNode struct{ ObjectID, ParentID, Name string }

type TemplateInstance struct {
	Nodes           []TemplateInstanceNode
	RootObjectID    string
	ObjectIDs       []string
	TemplateID      string
	TemplateVersion int64
}
