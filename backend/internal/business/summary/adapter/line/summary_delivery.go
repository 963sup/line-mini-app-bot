package line

import (
	summaryApplicationPort "github.com/963sup/line-mini-app-bot/internal/business/summary/application/port"
)

import (
	"context"
	"errors"

	lineplatform "github.com/963sup/line-mini-app-bot/internal/platform/line"
)

type SummaryDelivery struct {
	Bot *lineplatform.Messaging
}

func (d SummaryDelivery) Send(ctx context.Context, group, text, retry string) error {
	if e := d.Bot.PushText(ctx, group, text, retry); e != nil {
		if errors.Is(e, context.Canceled) {
			return context.Canceled
		}
		if errors.Is(e, context.DeadlineExceeded) {
			return context.DeadlineExceeded
		}
		return summaryApplicationPort.ErrDeliveryUnavailable
	}
	return nil
}
