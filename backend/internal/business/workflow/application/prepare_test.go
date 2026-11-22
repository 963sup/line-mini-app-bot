package application_test

import (
	"context"
	"errors"
	"reflect"
	"testing"

	agent "github.com/963sup/line-mini-app-bot/internal/business/agent/application"
	agentport "github.com/963sup/line-mini-app-bot/internal/business/agent/application/port"
	knowledge "github.com/963sup/line-mini-app-bot/internal/business/knowledge/application/port"
	operations "github.com/963sup/line-mini-app-bot/internal/business/operations/application"
	workflow "github.com/963sup/line-mini-app-bot/internal/business/workflow/application"
)

type extract func(context.Context, operations.Document) (operations.Extraction, error)

func (f extract) ExtractCandidates(c context.Context, d operations.Document) (operations.Extraction, error) {
	return f(c, d)
}

type rules func(context.Context, knowledge.Lookup) ([]knowledge.PublishedRule, error)

func (f rules) Retrieve(c context.Context, q knowledge.Lookup) ([]knowledge.PublishedRule, error) {
	return f(c, q)
}

type source struct{}

func (source) ReadPage(_ context.Context, actor, repository, revision string, page int) (string, error) {
	if actor != "user" || repository != "target" || revision != "v1" || page != 1 {
		return "", errors.New("unauthorized source")
	}
	return "Installed 2 units", nil
}

type model struct{}

func (model) Generate(context.Context, string, string) (string, error) {
	return `"Installed 2 units"`, nil
}

func TestPreparationComposesOptionalCapabilities(t *testing.T) {
	var stages []string
	var extractor operations.CandidateExtractor = extract(func(_ context.Context, doc operations.Document) (operations.Extraction, error) {
		stages = append(stages, "read")
		if doc.Filename == "denied" {
			return operations.Extraction{}, errors.New("denied")
		}
		candidates := []operations.Candidate{{Name: "installation"}}
		if doc.Filename == "complete" {
			value := "2"
			candidates[0].Quantity = &value
		}
		return operations.Extraction{Candidates: candidates}, nil
	})
	var reader knowledge.Reader = rules(func(_ context.Context, q knowledge.Lookup) ([]knowledge.PublishedRule, error) {
		stages = append(stages, "knowledge")
		if q.Actor != "user" {
			return nil, errors.New("unauthorized rule")
		}
		return []knowledge.PublishedRule{{ID: "rule", Version: "v1"}}, nil
	})
	var investigator agentport.Investigator = &agent.Investigator{Model: model{}, Pages: source{}}
	type plan = workflow.Preparation[operations.Document, operations.Extraction, []knowledge.PublishedRule, []agentport.Evidence, string]
	p := plan{
		Read: func(ctx context.Context, _ workflow.Scope, doc operations.Document) (operations.Extraction, error) {
			return extractor.ExtractCandidates(ctx, doc)
		},
		NeedKnowledge: func(d operations.Extraction) bool { return d.Candidates[0].Quantity == nil },
		Retrieve: func(ctx context.Context, scope workflow.Scope, _ operations.Document, _ operations.Extraction) ([]knowledge.PublishedRule, error) {
			return reader.Retrieve(ctx, knowledge.Lookup{Actor: scope.Actor, Organization: "org", TemplateVersion: "v1"})
		},
		NeedAgent: func(_ operations.Extraction, k *[]knowledge.PublishedRule) bool { return k != nil && len(*k) != 0 },
		Investigate: func(ctx context.Context, scope workflow.Scope, _ operations.Document, _ operations.Extraction, _ *[]knowledge.PublishedRule) ([]agentport.Evidence, error) {
			stages = append(stages, "agent")
			return investigator.Investigate(ctx, agentport.Investigation{Actor: scope.Actor, Repository: scope.Target, SourceRevision: "v1", AllowedPages: []int{1}, StepLimit: 1})
		},
		Build: func(_ operations.Extraction, k *[]knowledge.PublishedRule, e *[]agentport.Evidence) (string, error) {
			stages = append(stages, "build")
			if e == nil {
				return "direct", nil
			}
			return (*k)[0].Version + ": " + (*e)[0].Quote, nil
		},
	}
	for _, c := range []struct {
		file, result string
		stages       []string
	}{
		{"complete", "direct", []string{"read", "build"}},
		{"incomplete", "v1: Installed 2 units", []string{"read", "knowledge", "agent", "build"}},
		{"denied", "", []string{"read"}},
	} {
		t.Run(c.file, func(t *testing.T) {
			stages = nil
			out, err := workflow.Prepare(context.Background(), workflow.Scope{Actor: "user", Target: "target"}, operations.Document{Filename: c.file}, p)
			if (err != nil) != (c.file == "denied") || out != c.result || !reflect.DeepEqual(stages, c.stages) {
				t.Fatalf("result %q, error %v, steps %v", out, err, stages)
			}
		})
	}
}
