package bootstrap

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"

	"cloud.google.com/go/firestore"
	membershipdb "github.com/963sup/line-mini-app-bot/internal/business/membership/adapter/firestore"
	"github.com/963sup/line-mini-app-bot/internal/business/membership/application"
	"github.com/963sup/line-mini-app-bot/internal/platform/firestorekey"
)

func membershipInvitations(db *firestore.Client, channel, secret string) *application.Service {
	if db == nil || channel == "" || secret == "" {
		return nil
	}
	return &application.Service{
		Store: membershipdb.Store{DB: db, Channel: channel},
		Code: func(actor, event string) string {
			mac := hmac.New(sha256.New, []byte(secret))
			mac.Write([]byte(firestorekey.Encode("membership-invitation-v1", channel, actor, event)))
			return base64.RawURLEncoding.EncodeToString(mac.Sum(nil))
		},
	}
}
