package bootstrap

import (
	"fmt"
	"net/http"
	"os"
	"time"

	"cloud.google.com/go/firestore"
	groupdb "github.com/963sup/line-mini-app-bot/internal/linegroup/adapter/firestore"
	groupquery "github.com/963sup/line-mini-app-bot/internal/linegroup/application/query"
	lineplatform "github.com/963sup/line-mini-app-bot/internal/platform/line"
)

func newGroups(firestoreConfigured bool, getFirestore func() (*firestore.Client, error)) (*groupquery.Service, *groupdb.Store, error) {
	channel, token := os.Getenv("LINE_MESSAGING_API_CHANNEL_ID"), os.Getenv("LINE_MESSAGING_API_CHANNEL_ACCESS_TOKEN")
	if !firestoreConfigured || channel == "" || token == "" {
		return nil, nil, nil
	}
	client, e := getFirestore()
	if e != nil {
		return nil, nil, fmt.Errorf("cannot initialize group database")
	}
	store := &groupdb.Store{Client: client, ChannelID: channel}
	return &groupquery.Service{Store: store, Membership: &lineplatform.Membership{Token: token, Client: &http.Client{Timeout: time.Second, CheckRedirect: func(*http.Request, []*http.Request) error { return http.ErrUseLastResponse }}}}, store, nil
}
