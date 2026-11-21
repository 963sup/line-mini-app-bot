package domain

import (
	"crypto/sha256"
	"encoding/hex"
)

// MemberRef preserves the opaque organization-scoped ID exposed by People.
func MemberRef(organizationID, userID string) string {
	sum := sha256.Sum256([]byte(organizationID + ":" + userID))
	return hex.EncodeToString(sum[:])
}
