// Package line composes collection and summarization at the shared LINE entry point.
package line

import (
	summaryApplicationPortMessagecollection "github.com/963sup/line-mini-app-bot/internal/business/summary/application/port/messagecollection"
	summaryApplicationPortSummarization "github.com/963sup/line-mini-app-bot/internal/business/summary/application/port/summarization"
)

import (
	"context"

	collection "github.com/963sup/line-mini-app-bot/internal/business/summary/adapter/line/messagecollection"

	commands "github.com/963sup/line-mini-app-bot/internal/business/summary/adapter/line/summarization"
	transport "github.com/963sup/line-mini-app-bot/internal/transport/line"
	"github.com/line/line-bot-sdk-go/v8/linebot/webhook"
)

type Store interface {
	summaryApplicationPortMessagecollection.Store
	summaryApplicationPortSummarization.CommandStore
}

type Summary struct {
	Store Store
	Queue summaryApplicationPortSummarization.Queue
}

func (b *Summary) Observe(ctx context.Context, event webhook.EventInterface) (transport.Reply, bool, error) {
	if b == nil {
		return transport.Reply{}, false, nil
	}
	handler := commands.Commands{Store: b.Store, Queue: b.Queue}
	if reply, handled, err := handler.Reply(ctx, event); handled || err != nil {
		return reply, handled, err
	}
	collector := collection.Collection{Store: b.Store}
	return collector.Observe(ctx, event)
}
