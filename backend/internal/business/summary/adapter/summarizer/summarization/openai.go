package summarizer

import (
	summaryDomainEntityMessagecollection "github.com/963sup/line-mini-app-bot/internal/business/summary/domain/entity/messagecollection"
	summaryDomainEntitySummarization "github.com/963sup/line-mini-app-bot/internal/business/summary/domain/entity/summarization"
)

import (
	"context"

	"github.com/963sup/line-mini-app-bot/internal/genai"
)

type OpenAI genai.OpenAI

func (o *OpenAI) Summarize(ctx context.Context, messages []summaryDomainEntityMessagecollection.Message) (string, error) {
	input, _ := summaryDomainEntitySummarization.Input(messages)
	return (*genai.OpenAI)(o).Generate(ctx, summaryInstruction, input)
}
