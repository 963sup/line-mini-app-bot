package firestore

import (
	"context"

	"github.com/963sup/line-mini-app-bot/internal/business/summary/messagecollection"
	"github.com/963sup/line-mini-app-bot/internal/business/summary/summarization"
)

func (s *Summary) Edit(ctx context.Context, g, id, text string, timestamp int64) error {
	return translateError(s.editOperation(ctx, g, id, text, timestamp), nil)
}

func (s *Summary) Collect(ctx context.Context, g string, m messagecollection.Message) error {
	return translateError(s.collectOperation(ctx, g, m), nil)
}

func (s *Summary) Unsend(ctx context.Context, g, id string) error {
	return translateError(s.unsendOperation(ctx, g, id), nil)
}

func (s *Summary) Command(ctx context.Context, g, event, command string, timestamp int64) (summarization.Job, error) {
	value0, err := s.commandOperation(ctx, g, event, command, timestamp)
	return value0, translateError(err, nil)
}

func (s *Summary) Claim(ctx context.Context, g, id string) (summarization.Job, bool, error) {
	value0, value1, err := s.claimOperation(ctx, g, id)
	return value0, value1, translateError(err, nil)
}

func (s *Summary) Messages(ctx context.Context, j summarization.Job) ([]messagecollection.Message, error) {
	value0, err := s.messagesOperation(ctx, j)
	return value0, translateError(err, nil)
}

func (s *Summary) SaveResult(ctx context.Context, j summarization.Job, text string, sources []string) error {
	return translateError(s.saveResultOperation(ctx, j, text, sources), nil)
}

func (s *Summary) Done(ctx context.Context, j summarization.Job) error {
	return translateError(s.doneOperation(ctx, j), nil)
}

func (s *Summary) Current(ctx context.Context, j summarization.Job) (bool, error) {
	value0, err := s.currentOperation(ctx, j)
	return value0, translateError(err, nil)
}

func (s *Summary) Purge(ctx context.Context, j summarization.Job) error {
	return translateError(s.purgeOperation(ctx, j), nil)
}
