package operations

import (
	"context"

	source "github.com/963sup/line-mini-app-bot/internal/business/objectimport/domain/entity"
	automatic "github.com/963sup/line-mini-app-bot/internal/business/objectimport/domain/entity/automatic"
)

// Imports are the resource-authorized operations available to the import workflow.
type Imports interface {
	ImportTable(context.Context, string, string, string, string, []byte, map[string]string) (automatic.Run, error)
	Classify(context.Context, string, string, string, source.Source, map[string]string) (automatic.Run, error)
	Create(context.Context, string, string, string, source.Source, string) (automatic.Run, error)
}
