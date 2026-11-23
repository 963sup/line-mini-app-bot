package bootstrap

import (
	"errors"
	"net/http"
	"os"
	"regexp"
	"time"

	agentmodel "github.com/963sup/line-mini-app-bot/internal/business/agent/adapter/genai"
	agent "github.com/963sup/line-mini-app-bot/internal/business/agent/application"
	"github.com/963sup/line-mini-app-bot/internal/genai"
)

// NewInvestigator is opt-in: the caller supplies the authorized source reader.
func NewInvestigator(pages agent.PageReader) (*agent.Investigator, error) {
	model := os.Getenv("AGENT_MODEL")
	if pages == nil || !regexp.MustCompile(`^[A-Za-z0-9][A-Za-z0-9._-]*$`).MatchString(model) {
		return nil, errors.New("agent requires an authorized page reader and AGENT_MODEL")
	}
	client := &http.Client{Timeout: 45 * time.Second, CheckRedirect: func(*http.Request, []*http.Request) error { return http.ErrUseLastResponse }}
	var generator agentmodel.TextModel
	switch os.Getenv("AGENT_PROVIDER") {
	case "openai":
		key := os.Getenv("OPENAI_API_KEY")
		if key == "" {
			return nil, errors.New("OPENAI_API_KEY is missing")
		}
		generator = &genai.OpenAI{Key: key, Model: model, HTTP: client}
	case "gemini":
		key := os.Getenv("GEMINI_API_KEY")
		if key == "" {
			return nil, errors.New("GEMINI_API_KEY is missing")
		}
		generator = &genai.Gemini{Key: key, Model: model, HTTP: client}
	default:
		return nil, errors.New("AGENT_PROVIDER must be openai or gemini")
	}
	return &agent.Investigator{Model: agentmodel.EvidenceModel{Generator: generator}, Pages: pages}, nil
}
