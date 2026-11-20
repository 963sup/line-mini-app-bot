package server

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"time"

	"cloud.google.com/go/firestore"

	groupdb "github.com/963sup/line-mini-app-bot/internal/business/group/adapter/firestore"
	groupquery "github.com/963sup/line-mini-app-bot/internal/business/group/application/query"

	linetransport "github.com/963sup/line-mini-app-bot/internal/transport/line"
)

func newGroups(ctx context.Context) (*groupquery.Service, func() error, error) {
	cloud, channel, token := os.Getenv("GOOGLE_CLOUD_PROJECT"), os.Getenv("LINE_MESSAGING_API_CHANNEL_ID"), os.Getenv("LINE_MESSAGING_API_CHANNEL_ACCESS_TOKEN")
	if cloud == "" || channel == "" || token == "" {
		return nil, func() error { return nil }, nil
	}
	database := os.Getenv("FIRESTORE_DATABASE_ID")
	if database == "" {
		database = "(default)"
	}
	client, e := firestore.NewClientWithDatabase(ctx, cloud, database)
	if e != nil {
		return nil, nil, fmt.Errorf("cannot initialize group database")
	}
	return &groupquery.Service{Store: &groupdb.Store{Client: client, ChannelID: channel}, Membership: &linetransport.Membership{Token: token, Client: &http.Client{Timeout: time.Second, CheckRedirect: func(*http.Request, []*http.Request) error { return http.ErrUseLastResponse }}}}, client.Close, nil
}
