package port

import (
	"context"
	"errors"

	"github.com/963sup/line-mini-app-bot/internal/business/object/application/dto"
)

// CandidatePreparer receives only a scope already authorized by Object.
// It prepares read-only suggestions; Object owns any later create command.
type CandidatePreparer interface {
	PrepareCandidates(context.Context, string, string, dto.Document) (dto.CandidateResult, error)
}

// Preparation failures are expressed in the Object port contract.
var (
	ErrPreparationInvalidDocument = errors.New("object preparation: invalid document")
	ErrPreparationIncomplete      = errors.New("object preparation: incomplete response")
	ErrPreparationRefused         = errors.New("object preparation: request refused")
	ErrPreparationInvalidOutput   = errors.New("object preparation: invalid output")
	ErrPreparationProvider        = errors.New("object preparation: provider failed")
)
