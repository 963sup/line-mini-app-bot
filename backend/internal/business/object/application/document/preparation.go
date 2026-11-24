package document

import (
	"bytes"
	"context"
	"errors"

	"github.com/963sup/line-mini-app-bot/internal/business/object/application/dto"
	object "github.com/963sup/line-mini-app-bot/internal/business/object/application/port"
)

type Procedures interface {
	SuggestProcedures(context.Context, string, string) ([]Candidate, error)
}

type Preparation struct {
	Extract    CandidateExtractor
	Text       DocumentTextExtractor
	Procedures Procedures
}

var (
	_ object.CandidatePreparer = (*Preparation)(nil)
	_ object.TemplatePreparer  = (*Preparation)(nil)
)

func (b *Preparation) PrepareCandidates(ctx context.Context, actor, repository string, doc dto.Document) (dto.CandidateResult, error) {
	if actor == "" || repository == "" {
		return dto.CandidateResult{}, object.ErrPreparationInvalidDocument
	}
	return b.extract(ctx, doc)
}

func (b *Preparation) extract(ctx context.Context, doc dto.Document) (dto.CandidateResult, error) {
	if b == nil || b.Extract == nil {
		return dto.CandidateResult{}, object.ErrPreparationProvider
	}
	input := Document{Filename: doc.Filename, Data: doc.Data}
	var readText func(context.Context) (string, error)
	if b.Text != nil && bytes.HasPrefix(input.Data, []byte("%PDF-")) {
		readText = func(ctx context.Context) (string, error) { return b.Text.ExtractText(ctx, input) }
	}
	result, err := recognizeDocumentObjects(ctx, b.Extract, input, readText)
	return dto.CandidateResult{RawResponse: result.RawResponse, Candidates: mapCandidates(result.Candidates)}, preparationError(err)
}

func (b *Preparation) PrepareTemplate(ctx context.Context, actor string, in object.DraftPreparation) ([]dto.Candidate, error) {
	if in.Source == "procedure" {
		if b == nil || b.Procedures == nil {
			return nil, object.ErrPreparationProvider
		}
		if actor == "" {
			return nil, object.ErrPreparationInvalidDocument
		}
		result, err := b.Procedures.SuggestProcedures(ctx, in.Name, in.Conditions)
		if err != nil {
			return nil, preparationError(err)
		}
		return mapCandidates(result), nil
	}
	if actor == "" {
		return nil, object.ErrPreparationInvalidDocument
	}
	result, err := b.extract(ctx, dto.Document{Filename: in.Filename, Data: in.Data})
	return result.Candidates, err
}

func mapCandidates(items []Candidate) []dto.Candidate {
	result := make([]dto.Candidate, len(items))
	for i, candidate := range items {
		result[i] = dto.Candidate(candidate)
	}
	return result
}

func preparationError(err error) error {
	switch {
	case errors.Is(err, ErrInvalidDocument):
		return object.ErrPreparationInvalidDocument
	case errors.Is(err, ErrIncomplete):
		return object.ErrPreparationIncomplete
	case errors.Is(err, ErrRefused):
		return object.ErrPreparationRefused
	case errors.Is(err, ErrInvalidOutput):
		return object.ErrPreparationInvalidOutput
	case errors.Is(err, ErrProvider):
		return object.ErrPreparationProvider
	default:
		return err
	}
}
