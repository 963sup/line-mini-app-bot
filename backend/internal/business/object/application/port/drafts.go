package port

import (
	"context"

	operations "github.com/963sup/line-mini-app-bot/internal/business/operations/application"
)

// DraftPreparation is a read-only operation. Identity comes from the authenticated request.
type DraftPreparation struct {
	Source, Filename, Name, Conditions string
	Data                               []byte
}

// TemplatePreparer is owned by Object; bootstrap adapts the preparation workflow.
type TemplatePreparer interface {
	PrepareTemplate(context.Context, string, DraftPreparation) ([]operations.Candidate, error)
}
