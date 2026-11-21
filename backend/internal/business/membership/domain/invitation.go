package domain

import "time"

type Invitation struct {
	OrganizationID string
	IssuerID       string
	OwnerVersion   int64
	ExpiresAt      time.Time
	AcceptedBy     string
	AcceptedEvent  string
	Revoked        bool
}

func (i Invitation) CanAccept(ownerID string, ownerVersion int64, now time.Time) bool {
	return i.AcceptedBy == "" && !i.Revoked && ownerID != "" &&
		i.IssuerID == ownerID && i.OwnerVersion == ownerVersion && now.Before(i.ExpiresAt)
}
