// Package firestorekey preserves the shared persisted document ID encoding.
package firestorekey

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
)

func Encode(parts ...string) string {
	return JSONDigest(parts)
}

// JSONDigest preserves the digest used in persisted keys, receipts and cursors.
func JSONDigest(value any) string {
	b, _ := json.Marshal(value)
	return Digest(b)
}

// Digest hashes the exact bytes used by persisted keys, receipts and cursors.
func Digest(data []byte) string {
	sum := sha256.Sum256(data)
	return hex.EncodeToString(sum[:])
}
