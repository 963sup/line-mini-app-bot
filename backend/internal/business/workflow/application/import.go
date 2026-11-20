package workflow

import (
	"context"

	source "github.com/963sup/line-mini-app-bot/internal/business/objectimport/domain/entity"
	automatic "github.com/963sup/line-mini-app-bot/internal/business/objectimport/domain/entity/automatic"
	operations "github.com/963sup/line-mini-app-bot/internal/business/operations/application"
)

type ImportRequest struct {
	Source    source.Source
	TableFile []byte
	PerItem   bool
	BatchType string
	Decisions map[string]string
}

// Import routes new attempts and retries through the same resource-owned operations.
func Import(ctx context.Context, ops operations.Imports, actor, repository, key string, req ImportRequest) (automatic.Run, error) {
	if req.PerItem && req.BatchType != "" || !req.PerItem && req.Decisions != nil {
		return automatic.Run{}, source.ErrInput
	}
	if req.TableFile != nil {
		if !req.PerItem || req.Source.Pages != nil || req.Source.SourceRevision != "" {
			return automatic.Run{}, source.ErrInput
		}
		return ops.ImportTable(ctx, actor, repository, key, req.Source.SourceDocument, req.TableFile, req.Decisions)
	}
	if req.PerItem {
		return ops.Classify(ctx, actor, repository, key, req.Source, req.Decisions)
	}
	return ops.Create(ctx, actor, repository, key, req.Source, req.BatchType)
}
