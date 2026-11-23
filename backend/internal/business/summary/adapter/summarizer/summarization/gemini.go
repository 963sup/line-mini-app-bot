package summarizer

import (
	"context"

	"github.com/963sup/line-mini-app-bot/internal/genai"
)

type Gemini genai.Gemini

func (g *Gemini) Summarize(ctx context.Context, input string) (string, error) {
	return (*genai.Gemini)(g).Generate(ctx, summaryInstruction, input)
}
