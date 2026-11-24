package firestore_test

import (
	"errors"
	"os"
	"testing"
	"time"

	db "cloud.google.com/go/firestore"
	"github.com/963sup/line-mini-app-bot/internal/platform/identity"
	"github.com/963sup/line-mini-app-bot/internal/platform/identity/session"
	storage "github.com/963sup/line-mini-app-bot/internal/platform/identity/session/firestore"
)

func TestRevocationAndExpiry(t *testing.T) {
	if os.Getenv("FIRESTORE_EMULATOR_HOST") == "" {
		t.Skip("requires Firestore emulator")
	}
	ctx := t.Context()
	client, err := db.NewClient(ctx, "identity-session-test")
	if err != nil {
		t.Fatal(err)
	}
	defer client.Close()
	store := &storage.Store{Client: client}
	svc := session.New(store, "channel")
	now := time.Now().UTC().Truncate(time.Second)
	svc.Now = func() time.Time { return now }
	_, token, csrf, err := svc.Create(ctx, identity.User{ID: "actor", ExpiresAt: now.Add(time.Hour)}, "")
	if err != nil {
		t.Fatal(err)
	}
	if err := session.New(store, "other").Revoke(ctx, token, csrf); !errors.Is(err, session.ErrCSRF) {
		t.Fatalf("wrong channel: %v", err)
	}
	if err := svc.Revoke(ctx, token, token); !errors.Is(err, session.ErrCSRF) {
		t.Fatalf("wrong CSRF: %v", err)
	}
	now = now.Add(22 * time.Minute)
	renewed, err := svc.Refresh(ctx, token, csrf)
	if err != nil || renewed.Revision != 2 {
		t.Fatalf("refresh revision %d: %v", renewed.Revision, err)
	}
	now = now.Add(time.Hour)
	if _, err := svc.Load(ctx, token); !errors.Is(err, session.ErrInvalid) {
		t.Fatalf("expired session: %v", err)
	}
	if err := svc.Revoke(ctx, token, csrf); err != nil {
		t.Fatal(err)
	}
	if _, err := store.Get(ctx, session.Key(token)); !errors.Is(err, session.ErrInvalid) {
		t.Fatalf("revoked document: %v", err)
	}
	if err := svc.Revoke(ctx, token, csrf); err != nil {
		t.Fatalf("repeat revocation: %v", err)
	}
}
