package bootstrap

import (
	"context"
	"errors"
	"testing"

	agentmodel "github.com/963sup/line-mini-app-bot/internal/business/agent/adapter/genai"
	"github.com/963sup/line-mini-app-bot/internal/business/agent/application/port"
	"github.com/963sup/line-mini-app-bot/internal/genai"
)

type agentPages struct {
	calls int
	deny  bool
}

func (p *agentPages) ReadPage(_ context.Context, actor, repo, rev string, page int) (string, error) {
	p.calls++
	if p.deny || actor != "actor" || repo != "repo" || rev != "v1" || page != 7 {
		return "", errors.New("denied")
	}
	return "數量 20 SET", nil
}

type agentReply struct {
	value string
	calls int
}

func (r *agentReply) Generate(_ context.Context, _, _ string) (string, error) {
	r.calls++
	return r.value, nil
}

func TestAgentProviderAndEvidenceBoundary(t *testing.T) {
	for _, provider := range []string{"openai", "gemini"} {
		t.Run(provider, func(t *testing.T) {
			t.Setenv("AGENT_PROVIDER", provider)
			t.Setenv("AGENT_MODEL", "test-model")
			t.Setenv("OPENAI_API_KEY", "test")
			t.Setenv("GEMINI_API_KEY", "test")
			pages := &agentPages{}
			a, err := NewInvestigator(pages)
			if err != nil {
				t.Fatal(err)
			}
			switch a.Model.(agentmodel.EvidenceModel).Generator.(type) {
			case *genai.OpenAI:
				if provider != "openai" {
					t.Fatal(provider)
				}
			case *genai.Gemini:
				if provider != "gemini" {
					t.Fatal(provider)
				}
			default:
				t.Fatal("unexpected provider")
			}
			reply := &agentReply{value: "\"數量 20 SET\""}
			a.Model = agentmodel.EvidenceModel{Generator: reply}
			task := port.Investigation{Actor: "actor", Repository: "repo", SourceRevision: "v1", AllowedPages: []int{7, 8}, StepLimit: 1}
			evidence, err := a.Investigate(context.Background(), task)
			if err != nil || len(evidence) != 1 || evidence[0].Page != 7 || evidence[0].Quote != "數量 20 SET" {
				t.Fatal(evidence, err)
			}
			reply.value = "\"30 SET\""
			evidence, err = a.Investigate(context.Background(), task)
			if err != nil || len(evidence) != 0 || pages.calls != 2 {
				t.Fatal(evidence, err)
			}
			pages.deny = true
			_, err = a.Investigate(context.Background(), task)
			if err == nil || reply.calls != 2 {
				t.Fatal("model called after authorization failure")
			}
		})
	}
}
