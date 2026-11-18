// Package firestorekey preserves the shared persisted document ID encoding.
package firestorekey

import (
	"crypto/sha256"
	"encoding/json"
	"fmt"
)

func Encode(parts ...string) string {
	b, _ := json.Marshal(parts)
	return fmt.Sprintf("%x", sha256.Sum256(b))
}
