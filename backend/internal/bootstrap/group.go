package bootstrap

import (
	"fmt"

	"cloud.google.com/go/firestore"
	groupdb "github.com/963sup/line-mini-app-bot/internal/business/team/adapter/firestore/group"
	groupquery "github.com/963sup/line-mini-app-bot/internal/business/team/application/query"
	lineplatform "github.com/963sup/line-mini-app-bot/internal/platform/line"
)

func newGroups(firestoreConfigured bool, getFirestore func() (*firestore.Client, error), channel string, membership *lineplatform.Membership) (*groupquery.Service, *groupdb.Store, error) {
	if !firestoreConfigured || channel == "" || membership == nil {
		return nil, nil, nil
	}
	client, e := getFirestore()
	if e != nil {
		return nil, nil, fmt.Errorf("cannot initialize group database")
	}
	store := &groupdb.Store{Client: client, ChannelID: channel}
	return &groupquery.Service{Store: store, Membership: membership}, store, nil
}
