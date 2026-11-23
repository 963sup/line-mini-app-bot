package summarizer

import (
	"context"

	"github.com/963sup/line-mini-app-bot/internal/genai"
)

type OpenAI genai.OpenAI

func (o *OpenAI) Summarize(ctx context.Context, input string) (string, error) {
	return (*genai.OpenAI)(o).Generate(ctx, summaryInstruction, input)
}
