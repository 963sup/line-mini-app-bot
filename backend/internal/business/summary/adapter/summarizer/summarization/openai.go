package summarizer

import (
	"context"

	"github.com/963sup/line-mini-app-bot/internal/platform/textgeneration"
)

type OpenAI textgeneration.OpenAI

func (o *OpenAI) Summarize(ctx context.Context, input string) (string, error) {
	return (*textgeneration.OpenAI)(o).Generate(ctx, summaryInstruction, input)
}
