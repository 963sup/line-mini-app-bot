package summary

import (
	"context"
	"errors"
	"net/http"
	"time"

	openai "github.com/sashabaranov/go-openai"
)

// OpenAI implements the existing model boundary using Chat Completions.
type OpenAI struct {
	Key, Model string
	HTTP       *http.Client
}

func (o *OpenAI) Summarize(ctx context.Context, messages []Message) (string, error) {
	ctx, cancel := context.WithTimeout(ctx, 45*time.Second)
	defer cancel()
	input, sources := Input(messages)
	config := openai.DefaultConfig(o.Key)
	if o.HTTP != nil {
		config.HTTPClient = o.HTTP
	}
	response, err := openai.NewClientWithConfig(config).CreateChatCompletion(ctx, openai.ChatCompletionRequest{
		Model:               o.Model,
		MaxCompletionTokens: 1500,
		Messages: []openai.ChatCompletionMessage{
			{Role: openai.ChatMessageRoleSystem, Content: summaryInstruction},
			{Role: openai.ChatMessageRoleUser, Content: input},
		},
	})
	if err != nil {
		// SDK errors may contain provider response details; never expose them to LINE.
		return "", errors.New("summary model unavailable")
	}
	if len(response.Choices) != 1 || response.Choices[0].FinishReason != openai.FinishReasonStop || response.Choices[0].Message.Refusal != "" || len(response.Choices[0].Message.ToolCalls) != 0 {
		return "", errors.New("model did not finish summary")
	}
	return validateSummary(response.Choices[0].Message.Content, len(sources))
}
