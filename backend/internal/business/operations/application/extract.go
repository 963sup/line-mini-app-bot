package operations

import (
	"context"

	parsing "github.com/963sup/line-mini-app-bot/internal/business/objectimport/domain/entity/parsing"
)

// ExtractCandidates delegates to the resource owner's authorized parsing use case.
type ExtractCandidates func(context.Context, string, string, parsing.Input) (parsing.Result, error)
