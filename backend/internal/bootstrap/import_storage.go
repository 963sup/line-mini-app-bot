package bootstrap

import (
	"context"

	db "cloud.google.com/go/firestore"
	knowledge "github.com/963sup/line-mini-app-bot/internal/business/knowledge/adapter/firestore"
	rules "github.com/963sup/line-mini-app-bot/internal/business/knowledge/application/port"
	blobs "github.com/963sup/line-mini-app-bot/internal/business/objectimport/adapter/cloudstorage"
	ports "github.com/963sup/line-mini-app-bot/internal/business/objectimport/application/port/automatic"
	storage "google.golang.org/api/storage/v1"
)

type ImportStorage struct {
	Knowledge rules.Reader
	Sources   ports.SourceStore
}

func NewImportStorage(ctx context.Context, client *db.Client, channel, bucket string) (ImportStorage, error) {
	out := ImportStorage{Knowledge: knowledge.Reader{DB: client, Channel: channel}}
	if bucket != "" {
		gcs, err := storage.NewService(ctx)
		if err != nil {
			return ImportStorage{}, err
		}
		out.Sources = blobs.Sources{Client: gcs, Bucket: bucket, Channel: channel}
	}
	return out, nil
}
