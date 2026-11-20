package firestorekey

import (
	"encoding/base64"
	"encoding/json"
)

type cursor struct{ Scope, Last string }

// EncodeCursor binds the last document ID to the scope of a paginated query.
func EncodeCursor(scope any, last string) string {
	b, _ := json.Marshal(cursor{JSONDigest(scope), last})
	return base64.RawURLEncoding.EncodeToString(b)
}

// DecodeCursor verifies the scope. The caller validates the document ID for its query.
func DecodeCursor(token string, scope any) (string, bool) {
	b, err := base64.RawURLEncoding.DecodeString(token)
	var c cursor
	if err != nil || json.Unmarshal(b, &c) != nil || c.Scope != JSONDigest(scope) {
		return "", false
	}
	return c.Last, true
}
