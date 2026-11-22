package bootstrap

import (
	"context"

	object "github.com/963sup/line-mini-app-bot/internal/business/object/application"
	operations "github.com/963sup/line-mini-app-bot/internal/business/operations/application"
	workflow "github.com/963sup/line-mini-app-bot/internal/business/workflow/application"
)

type draftOperations interface {
	operations.CandidateExtractor
	SuggestProcedures(context.Context, string, string) ([]operations.Candidate, error)
}

type templatePreparation struct{ operations draftOperations }

var _ object.TemplatePreparer = templatePreparation{}

// This bounded workflow needs neither dynamic Agent investigation nor Knowledge retrieval.
// Object owns the resulting tree and validation; no preparation step writes data.
func (p templatePreparation) PrepareTemplate(ctx context.Context, actor string, in object.DraftPreparation) ([]operations.Candidate, error) {
	return workflow.Prepare(ctx, workflow.Scope{Actor: actor, Target: "personal-template"}, in,
		workflow.Preparation[object.DraftPreparation, []operations.Candidate, struct{}, struct{}, []operations.Candidate]{
			Read: func(ctx context.Context, _ workflow.Scope, in object.DraftPreparation) ([]operations.Candidate, error) {
				if in.Source == "procedure" {
					return p.operations.SuggestProcedures(ctx, in.Name, in.Conditions)
				}
				result, err := p.operations.ExtractCandidates(ctx, operations.Document{Filename: in.Filename, Data: in.Data})
				return result.Candidates, err
			},
			Build: func(c []operations.Candidate, _ *struct{}, _ *struct{}) ([]operations.Candidate, error) {
				return c, nil
			},
		})
}
