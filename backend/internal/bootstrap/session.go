package bootstrap

import (
	"context"
	"os"

	"cloud.google.com/go/firestore"
	"github.com/963sup/line-mini-app-bot/internal/identity/session"
	storage "github.com/963sup/line-mini-app-bot/internal/identity/session/firestore"
)

func newSessions(ctx context.Context, channel string) (*session.Service, *storage.Store, func() error, error) {
	project := os.Getenv("GOOGLE_CLOUD_PROJECT")
	if project == "" || channel == "" {
		return nil, nil, func() error { return nil }, nil
	}
	database := os.Getenv("FIRESTORE_DATABASE_ID")
	if database == "" {
		database = "(default)"
	}
	client, err := firestore.NewClientWithDatabase(ctx, project, database)
	if err != nil {
		return nil, nil, nil, err
	}
	store := &storage.Store{Client: client}
	return &session.Service{Store: store, ChannelID: channel}, store, client.Close, nil
}
