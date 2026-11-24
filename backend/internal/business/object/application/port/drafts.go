package port

import (
	"context"
	"errors"

	"github.com/963sup/line-mini-app-bot/internal/business/object/application/dto"
)

var ErrInvalidCandidate = errors.New("no usable object candidates")

// DraftPreparation is a read-only operation. Identity comes from the authenticated request.
type DraftPreparation struct {
	Source, Filename, Name, Conditions string
	Data                               []byte
}

// TemplatePreparer is owned by Object; bootstrap adapts the read-only preparation operation.
type TemplatePreparer interface {
	PrepareTemplate(context.Context, string, DraftPreparation) ([]dto.Candidate, error)
}
