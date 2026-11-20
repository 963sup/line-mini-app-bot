package summarizer

import (
	summaryDomainEntityMessagecollection "github.com/963sup/line-mini-app-bot/internal/business/summary/domain/entity/messagecollection"
	summaryDomainEntitySummarization "github.com/963sup/line-mini-app-bot/internal/business/summary/domain/entity/summarization"
)

import (
	"context"
	"errors"
	"net/http"
	"time"

	"github.com/openai/openai-go/v3"
	"github.com/openai/openai-go/v3/option"
)

// OpenAI implements the existing model boundary using Chat Completions.
type OpenAI struct {
	Key, Model string
	HTTP       *http.Client
}

func (o *OpenAI) Summarize(ctx context.Context, messages []summaryDomainEntityMessagecollection.Message) (string, error) {
	ctx, cancel := context.WithTimeout(ctx, 45*time.Second)
	defer cancel()
	input, _ := summaryDomainEntitySummarization.Input(messages)
	options := []option.RequestOption{
		option.WithAPIKey(o.Key),
		option.WithBaseURL("https://api.openai.com/v1/"),
		option.WithMaxRetries(0),
	}
	if o.HTTP != nil {
		options = append(options, option.WithHTTPClient(o.HTTP))
	}
	provider := openai.NewClient(options...)
	response, err := provider.Chat.Completions.New(ctx, openai.ChatCompletionNewParams{
		Model:               o.Model,
		MaxCompletionTokens: openai.Int(1500),
		Messages: []openai.ChatCompletionMessageParamUnion{
			openai.SystemMessage(summaryInstruction),
			openai.UserMessage(input),
		},
	})
	if err != nil {
		// SDK errors may contain provider response details; never expose them to LINE.
		return "", errors.New("summary model unavailable")
	}
	if len(response.Choices) != 1 || response.Choices[0].FinishReason != "stop" || response.Choices[0].Message.Refusal != "" || len(response.Choices[0].Message.ToolCalls) != 0 {
		return "", errors.New("model did not finish summary")
	}
	return response.Choices[0].Message.Content, nil
}
