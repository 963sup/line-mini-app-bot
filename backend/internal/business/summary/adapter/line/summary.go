// Package line composes collection and summarization at the shared LINE entry point.
package line

import (
	summaryApplicationPortMessagecollection "github.com/963sup/line-mini-app-bot/internal/business/summary/application/port/messagecollection"
	summaryApplicationPortSummarization "github.com/963sup/line-mini-app-bot/internal/business/summary/application/port/summarization"
)

type Store interface {
	summaryApplicationPortMessagecollection.Store
	summaryApplicationPortSummarization.CommandStore
}

type Summary struct {
	Store Store
	Queue summaryApplicationPortSummarization.Queue
}
