package summarization

import (
	"context"

	summaryDomainEntityMessagecollection "github.com/963sup/line-mini-app-bot/internal/business/summary/domain/entity/messagecollection"
	summaryDomainEntitySummarization "github.com/963sup/line-mini-app-bot/internal/business/summary/domain/entity/summarization"
)

type CommandStore interface {
	Command(context.Context, string, string, string, int64) (summaryDomainEntitySummarization.Job, error)
}

type Store interface {
	Claim(context.Context, string, string) (summaryDomainEntitySummarization.Job, bool, error)
	Messages(context.Context, summaryDomainEntitySummarization.Job) ([]summaryDomainEntityMessagecollection.Message, error)
	SaveResult(context.Context, summaryDomainEntitySummarization.Job, string, []string) error
	Current(context.Context, summaryDomainEntitySummarization.Job) (bool, error)
	Done(context.Context, summaryDomainEntitySummarization.Job) error
	Purge(context.Context, summaryDomainEntitySummarization.Job) error
}

type Model interface {
	Summarize(context.Context, []summaryDomainEntityMessagecollection.Message) (string, error)
}

type Queue interface {
	Enqueue(context.Context, summaryDomainEntitySummarization.Job) error
}

type Sender interface {
	Send(context.Context, string, string, string) error
}
