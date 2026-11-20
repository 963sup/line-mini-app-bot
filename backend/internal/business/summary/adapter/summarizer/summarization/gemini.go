package summarizer

import (
	summaryDomainEntityMessagecollection "github.com/963sup/line-mini-app-bot/internal/business/summary/domain/entity/messagecollection"
	summaryDomainEntitySummarization "github.com/963sup/line-mini-app-bot/internal/business/summary/domain/entity/summarization"
)

import (
	"context"

	"github.com/963sup/line-mini-app-bot/internal/genai"
)

type Gemini genai.Gemini

func (g *Gemini) Summarize(ctx context.Context, messages []summaryDomainEntityMessagecollection.Message) (string, error) {
	input, _ := summaryDomainEntitySummarization.Input(messages)
	return (*genai.Gemini)(g).Generate(ctx, summaryInstruction, input)
}
