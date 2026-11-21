package bootstrap

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"time"

	"cloud.google.com/go/firestore"
	groupdb "github.com/963sup/line-mini-app-bot/internal/business/group/adapter/firestore"
	groupquery "github.com/963sup/line-mini-app-bot/internal/business/group/application/query"
	lineplatform "github.com/963sup/line-mini-app-bot/internal/platform/line"
)

func newGroups(ctx context.Context) (*groupquery.Service, *groupdb.Store, func() error, error) {
	cloud, channel, token := os.Getenv("GOOGLE_CLOUD_PROJECT"), os.Getenv("LINE_MESSAGING_API_CHANNEL_ID"), os.Getenv("LINE_MESSAGING_API_CHANNEL_ACCESS_TOKEN")
	if cloud == "" || channel == "" || token == "" {
		return nil, nil, func() error { return nil }, nil
	}
	database := os.Getenv("FIRESTORE_DATABASE_ID")
	if database == "" {
		database = "(default)"
	}
	client, e := firestore.NewClientWithDatabase(ctx, cloud, database)
	if e != nil {
		return nil, nil, nil, fmt.Errorf("cannot initialize group database")
	}
	store := &groupdb.Store{Client: client, ChannelID: channel}
	return &groupquery.Service{Store: store, Membership: &lineplatform.Membership{Token: token, Client: &http.Client{Timeout: time.Second, CheckRedirect: func(*http.Request, []*http.Request) error { return http.ErrUseLastResponse }}}}, store, client.Close, nil
}
