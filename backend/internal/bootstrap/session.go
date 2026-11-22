package bootstrap

import (
	"cloud.google.com/go/firestore"
	"github.com/963sup/line-mini-app-bot/internal/identity/session"
	storage "github.com/963sup/line-mini-app-bot/internal/identity/session/firestore"
)

func newSessions(firestoreConfigured bool, getFirestore func() (*firestore.Client, error), channel string) (*session.Service, *storage.Store, error) {
	if !firestoreConfigured || channel == "" {
		return nil, nil, nil
	}
	client, err := getFirestore()
	if err != nil {
		return nil, nil, err
	}
	store := &storage.Store{Client: client}
	return &session.Service{Store: store, ChannelID: channel}, store, nil
}
