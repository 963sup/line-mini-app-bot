package bootstrap

import (
	"bytes"
	"context"
	"errors"
	"strings"

	agent "github.com/963sup/line-mini-app-bot/internal/business/agent/application"
	knowledge "github.com/963sup/line-mini-app-bot/internal/business/knowledge/application"
	"github.com/963sup/line-mini-app-bot/internal/business/object/application/dto"
	object "github.com/963sup/line-mini-app-bot/internal/business/object/application/port"
	operations "github.com/963sup/line-mini-app-bot/internal/business/operations/application"
	workflow "github.com/963sup/line-mini-app-bot/internal/business/workflow/application"
)

type Procedures interface {
	SuggestProcedures(context.Context, string, string) ([]operations.Candidate, error)
}

type objectPreparation struct {
	Extract    operations.CandidateExtractor
	Text       operations.DocumentTextExtractor
	Procedures Procedures
}

var (
	_ object.CandidatePreparer = (*objectPreparation)(nil)
	_ object.TemplatePreparer  = (*objectPreparation)(nil)
)

func (b *objectPreparation) PrepareCandidates(ctx context.Context, actor, repository string, doc dto.Document) (dto.CandidateResult, error) {
	if actor == "" || repository == "" {
		return dto.CandidateResult{}, object.ErrPreparationInvalidDocument
	}
	return b.extract(ctx, doc)
}

func (b *objectPreparation) extract(ctx context.Context, doc dto.Document) (dto.CandidateResult, error) {
	if b == nil || b.Extract == nil {
		return dto.CandidateResult{}, object.ErrPreparationProvider
	}
	input := operations.Document{Filename: doc.Filename, Data: doc.Data}
	var readText func(context.Context) (string, error)
	if b.Text != nil && bytes.HasPrefix(input.Data, []byte("%PDF-")) {
		readText = func(ctx context.Context) (string, error) { return b.Text.ExtractText(ctx, input) }
	}
	result, err := workflow.PrepareDocument(ctx,
		func(ctx context.Context) (dto.CandidateResult, error) {
			result, err := b.Extract.ExtractCandidates(ctx, input)
			return dto.CandidateResult{RawResponse: result.RawResponse, Candidates: mapCandidates(result.Candidates)}, err
		},
		func(result dto.CandidateResult) []dto.Candidate { return result.Candidates },
		func(candidate dto.Candidate) string {
			if candidate.Location != nil && strings.TrimSpace(*candidate.Location) != "" {
				return ""
			}
			return strings.TrimSpace(candidate.Quote)
		},
		readText, knowledge.IndexPages,
		func(candidate *dto.Candidate, pages []int) {
			if location := agent.InferLocation(pages); location != nil {
				candidate.Location = location
			}
		},
	)
	return result, preparationError(err)
}

func (b *objectPreparation) PrepareTemplate(ctx context.Context, actor string, in object.DraftPreparation) ([]dto.Candidate, error) {
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

func mapCandidates(items []operations.Candidate) []dto.Candidate {
	result := make([]dto.Candidate, len(items))
	for i, candidate := range items {
		result[i] = dto.Candidate(candidate)
	}
	return result
}

func preparationError(err error) error {
	switch {
	case errors.Is(err, operations.ErrInvalidDocument):
		return object.ErrPreparationInvalidDocument
	case errors.Is(err, operations.ErrIncomplete):
		return object.ErrPreparationIncomplete
	case errors.Is(err, operations.ErrRefused):
		return object.ErrPreparationRefused
	case errors.Is(err, operations.ErrInvalidOutput):
		return object.ErrPreparationInvalidOutput
	case errors.Is(err, operations.ErrProvider):
		return object.ErrPreparationProvider
	default:
		return err
	}
}
