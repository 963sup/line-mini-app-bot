package bootstrap

import (
	"context"
	"errors"
	"testing"

	operations "github.com/963sup/line-mini-app-bot/internal/business/object/application/document"
	"github.com/963sup/line-mini-app-bot/internal/business/object/application/dto"
	object "github.com/963sup/line-mini-app-bot/internal/business/object/application/port"
)

type extraction func(context.Context, operations.Document) (operations.Extraction, error)

func (f extraction) ExtractCandidates(ctx context.Context, d operations.Document) (operations.Extraction, error) {
	return f(ctx, d)
}

type documentText func(context.Context, operations.Document) (string, error)

func (f documentText) ExtractText(ctx context.Context, doc operations.Document) (string, error) {
	return f(ctx, doc)
}

func TestDocumentEntrypointsReturnOnlyUnambiguousSourceLocations(t *testing.T) {
	provided := "原本位置"
	textCalls := 0
	bridge := &operations.Preparation{
		Extract: extraction(func(context.Context, operations.Document) (operations.Extraction, error) {
			return operations.Extraction{Candidates: []operations.Candidate{
				{Name: "Pump", Quote: "Pump 12 units"},
				{Name: "Ambiguous", Quote: "repeated"},
				{Name: "Missing", Quote: "missing"},
				{Name: "Existing", Quote: "Pump 12 units", Location: &provided},
			}}, nil
		}),
		Text: documentText(func(_ context.Context, doc operations.Document) (string, error) {
			textCalls++
			if string(doc.Data) != "%PDF-test" {
				t.Fatal("evidence reader did not receive the submitted file")
			}
			return "repeated\fPump\n 12 units repeated\f", nil
		}),
	}
	doc := dto.Document{Filename: "estimate.pdf", Data: []byte("%PDF-test")}
	result, err := bridge.PrepareCandidates(t.Context(), "actor", "repository", doc)
	if err != nil {
		t.Fatal(err)
	}
	draft, err := bridge.PrepareTemplate(t.Context(), "actor", object.DraftPreparation{Source: "document", Filename: doc.Filename, Data: doc.Data})
	if err != nil {
		t.Fatal(err)
	}
	for _, items := range [][]dto.Candidate{result.Candidates, draft} {
		if len(items) != 4 || items[0].Location == nil || *items[0].Location != "第 2 頁" ||
			items[1].Location != nil || items[2].Location != nil || items[3].Location != &provided {
			t.Fatalf("unexpected reviewed candidates: %+v", items)
		}
	}
	if textCalls != 2 {
		t.Fatalf("expected one source read per request, got %d", textCalls)
	}
}

func TestObjectPreparationReturnsOnlyDeliveredCandidates(t *testing.T) {
	var calls int
	b := &operations.Preparation{Extract: extraction(func(_ context.Context, doc operations.Document) (operations.Extraction, error) {
		calls++
		if doc.Filename != "estimate.pdf" || string(doc.Data) != "source" {
			t.Fatalf("lost document: %+v", doc)
		}
		return operations.Extraction{Candidates: []operations.Candidate{{Name: "item", Quote: "item"}}, RawResponse: []byte("response")}, nil
	})}
	doc := dto.Document{Filename: "estimate.pdf", Data: []byte("source")}
	result, err := b.PrepareCandidates(t.Context(), "actor", "repository", doc)
	if err != nil || len(result.Candidates) != 1 || result.Candidates[0].Name != "item" || string(result.RawResponse) != "response" || calls != 1 {
		t.Fatalf("result=%+v err=%v calls=%d", result, err, calls)
	}
	items, err := b.PrepareTemplate(t.Context(), "actor", object.DraftPreparation{Source: "document", Filename: doc.Filename, Data: doc.Data})
	if err != nil || len(items) != 1 || items[0].Name != "item" || calls != 2 {
		t.Fatalf("template=%+v err=%v calls=%d", items, err, calls)
	}
	if _, err := b.PrepareCandidates(t.Context(), "", "repository", doc); err == nil || calls != 2 {
		t.Fatalf("missing actor reached extractor: %v calls=%d", err, calls)
	}
	if _, err := b.PrepareCandidates(t.Context(), "actor", "", doc); err == nil || calls != 2 {
		t.Fatalf("missing target reached extractor: %v calls=%d", err, calls)
	}
	if _, err := b.PrepareTemplate(t.Context(), "", object.DraftPreparation{Source: "document", Filename: doc.Filename, Data: doc.Data}); err == nil || calls != 2 {
		t.Fatalf("missing template actor reached extractor: %v calls=%d", err, calls)
	}
}

func TestObjectPreparationPreservesProviderFailureDetails(t *testing.T) {
	b := &operations.Preparation{Extract: extraction(func(context.Context, operations.Document) (operations.Extraction, error) {
		return operations.Extraction{RawResponse: []byte("incomplete")}, operations.ErrIncomplete
	})}
	result, err := b.PrepareCandidates(t.Context(), "actor", "repository", dto.Document{Filename: "estimate.pdf"})
	if !errors.Is(err, object.ErrPreparationIncomplete) || string(result.RawResponse) != "incomplete" {
		t.Fatalf("result=%+v err=%v", result, err)
	}
}
