package summarizer

import (
	"context"

	"github.com/963sup/line-mini-app-bot/internal/platform/textgeneration"
)

type Gemini textgeneration.Gemini

func (g *Gemini) Summarize(ctx context.Context, input string) (string, error) {
	return (*textgeneration.Gemini)(g).Generate(ctx, summaryInstruction, input)
}
