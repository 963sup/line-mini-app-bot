package application

import (
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/963sup/line-mini-app-bot/internal/business/agent/application/port"
)

type quoteModel func(context.Context, string) (*string, error)

func (f quoteModel) Quote(ctx context.Context, source string) (*string, error) { return f(ctx, source) }

type pageReader func(context.Context, string, string, string, int) (string, error)

func (f pageReader) ReadPage(ctx context.Context, actor, repo, revision string, page int) (string, error) {
	return f(ctx, actor, repo, revision, page)
}

func TestInvestigationTrustAndBudget(t *testing.T) {
	value := func(s string) *string { return &s }
	denied, unavailable := errors.New("denied"), errors.New("model unavailable")
	for _, tc := range []struct {
		name                       string
		replies                    []*string
		limit, page, reads, models int
		readError, modelError      error
		large                      bool
	}{
		{name: "literal", replies: []*string{value("20 SET")}, limit: 1, page: 7, reads: 1, models: 1},
		{name: "fabricated then literal", replies: []*string{value("30 SET"), value("20 SET")}, limit: 2, page: 8, reads: 2, models: 2},
		{name: "absent then literal", replies: []*string{nil, value("20 SET")}, limit: 2, page: 8, reads: 2, models: 2},
		{name: "budget exhausted", replies: []*string{nil, value("20 SET")}, limit: 1, reads: 1, models: 1},
		{name: "blank is not evidence", replies: []*string{value(" "), value("")}, limit: 2, reads: 2, models: 2},
		{name: "authorization denied", limit: 1, readError: denied, reads: 1},
		{name: "provider failure", limit: 1, modelError: unavailable, reads: 1, models: 1},
		{name: "oversize page", limit: 1, large: true, reads: 1},
	} {
		t.Run(tc.name, func(t *testing.T) {
			reads, models := 0, 0
			s := Investigator{
				Pages: pageReader(func(_ context.Context, actor, repo, revision string, page int) (string, error) {
					reads++
					if actor != "actor" || repo != "repo" || revision != "v1" || page != 6+reads {
						t.Fatal("source scope changed")
					}
					if tc.large {
						return strings.Repeat("x", (32<<10)+1), nil
					}
					return "數量 20 SET", tc.readError
				}),
				Model: quoteModel(func(_ context.Context, source string) (*string, error) {
					models++
					if source != "數量 20 SET" {
						t.Fatal(source)
					}
					if tc.modelError != nil {
						return nil, tc.modelError
					}
					return tc.replies[models-1], nil
				}),
			}
			task := port.Investigation{Actor: "actor", Repository: "repo", SourceRevision: "v1", AllowedPages: []int{7, 8}, StepLimit: tc.limit}
			evidence, err := s.Investigate(t.Context(), task)
			if (err != nil) != (tc.readError != nil || tc.modelError != nil || tc.large) {
				t.Fatal(evidence, err)
			}
			if tc.readError != nil && !errors.Is(err, tc.readError) || tc.modelError != nil && !errors.Is(err, tc.modelError) {
				t.Fatal(err)
			}
			if reads != tc.reads || models != tc.models {
				t.Fatal("budget or authorization bypassed", reads, models)
			}
			if tc.page == 0 {
				if len(evidence) != 0 {
					t.Fatal("unverified evidence returned", evidence)
				}
			} else if len(evidence) != 1 || evidence[0].Page != tc.page || evidence[0].Quote != "20 SET" {
				t.Fatal(evidence)
			}
		})
	}
}

func TestInvalidInvestigationDoesNotReadOrGenerate(t *testing.T) {
	for _, task := range []port.Investigation{
		{},
		{Actor: "actor", Repository: "repo", SourceRevision: "v1", AllowedPages: []int{0}, StepLimit: 1},
		{Actor: "actor", Repository: "repo", SourceRevision: "v1", AllowedPages: []int{1}, StepLimit: 0},
	} {
		// Missing dependencies make any I/O before validation fail this test.
		s := Investigator{}
		if _, err := s.Investigate(t.Context(), task); err == nil {
			t.Fatal("invalid scope accepted")
		}
	}
}
