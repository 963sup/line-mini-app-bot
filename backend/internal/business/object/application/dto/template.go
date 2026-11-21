package dto

import (
	"time"

	"github.com/963sup/line-mini-app-bot/internal/business/object/domain/template"
)

// Template is an immutable, immediately usable v1 snapshot.
type Template struct {
	ID, RepositoryID, Name, CreatorID string
	Version                           int64
	Nodes                             []template.Node
	CreatedAt                         time.Time
}

type TemplateCreation struct {
	Name  string
	Nodes []template.Node
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
