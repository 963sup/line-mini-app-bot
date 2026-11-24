package bootstrap

import (
	"context"
	"errors"
	"testing"

	object "github.com/963sup/line-mini-app-bot/internal/business/object/application"
	operations "github.com/963sup/line-mini-app-bot/internal/business/operations/application"
)

type draftOps struct {
	documents, procedures int
	err                   error
}

func (p *draftOps) ExtractCandidates(_ context.Context, in operations.Document) (operations.Extraction, error) {
	p.documents++
	if in.Filename != "estimate.pdf" {
		panic("lost document")
	}
	return operations.Extraction{Candidates: []operations.Candidate{{Name: "來源工項"}}}, p.err
}

func (p *draftOps) SuggestProcedures(_ context.Context, name, conditions string) ([]operations.Candidate, error) {
	p.procedures++
	if name != "防水" || conditions != "室內" {
		panic("lost conditions")
	}
	return []operations.Candidate{{Name: "建議工序"}}, p.err
}

func TestTemplatePreparationDispatch(t *testing.T) {
	ops := &draftOps{}
	p := &objectPreparation{Extract: ops, Procedures: ops}
	doc, err := p.PrepareTemplate(t.Context(), "A", object.DraftPreparation{Source: "document", Filename: "estimate.pdf"})
	if err != nil || len(doc) != 1 || doc[0].Name != "來源工項" || ops.documents != 1 || ops.procedures != 0 {
		t.Fatal(doc, err)
	}
	steps, err := p.PrepareTemplate(t.Context(), "A", object.DraftPreparation{Source: "procedure", Name: "防水", Conditions: "室內"})
	if err != nil || len(steps) != 1 || steps[0].Name != "建議工序" || ops.documents != 1 || ops.procedures != 1 {
		t.Fatal(steps, err)
	}
	ops.err = context.Canceled
	if _, err := p.PrepareTemplate(t.Context(), "A", object.DraftPreparation{Source: "procedure", Name: "防水", Conditions: "室內"}); !errors.Is(err, context.Canceled) {
		t.Fatal(err)
	}
	if _, err := p.PrepareTemplate(t.Context(), "", object.DraftPreparation{Source: "procedure"}); err == nil {
		t.Fatal("missing identity accepted")
	}
}
