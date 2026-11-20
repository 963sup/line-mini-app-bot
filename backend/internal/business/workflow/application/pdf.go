package workflow

import (
	"context"

	parsing "github.com/963sup/line-mini-app-bot/internal/business/objectimport/domain/entity/parsing"
	operations "github.com/963sup/line-mini-app-bot/internal/business/operations/application"
)

// PDF prepares candidates for review. Confirmation remains a separate operation.
type PDF struct {
	Extract operations.ExtractCandidates
}

func (w PDF) Parse(ctx context.Context, actor, repository string, input parsing.Input) (parsing.Result, error) {
	return w.Extract(ctx, actor, repository, input)
}
