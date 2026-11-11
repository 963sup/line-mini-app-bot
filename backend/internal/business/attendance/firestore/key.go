package firestore

import "github.com/963sup/line-mini-app-bot/internal/persistence/firestorekey"

func documentKey(parts ...string) string { return firestorekey.Encode(parts...) }
