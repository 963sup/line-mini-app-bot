package bootstrap

import (
	"cloud.google.com/go/firestore"
	"github.com/963sup/line-mini-app-bot/internal/platform/identity/session"
	storage "github.com/963sup/line-mini-app-bot/internal/platform/identity/session/firestore"
)

func newSessions(firestoreConfigured bool, getFirestore func() (*firestore.Client, error), channel string) (*session.Service, error) {
	if !firestoreConfigured || channel == "" {
		return nil, nil
	}
	client, err := getFirestore()
	if err != nil {
		return nil, err
	}
	store := &storage.Store{Client: client}
	return session.New(store, channel), nil
}
