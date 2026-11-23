package entity

import "slices"

type Organization struct {
	Name           string
	ID             string
	OwnerUserID    string
	MemberUserIDs  []string
	Disabled       bool
	MembershipMode string
}

// LegacyEnabled prevents legacy fields from authorizing a resource after it
// switches to formal membership. Existing documents have no mode field.
func (o Organization) LegacyEnabled() bool {
	return !o.Disabled && o.OwnerUserID != "" && (o.MembershipMode == "" || o.MembershipMode == "legacy")
}

func (o Organization) VisibleLegacy(actor string) bool {
	return actor != "" && o.LegacyEnabled() && (o.OwnerUserID == actor || slices.Contains(o.MemberUserIDs, actor))
}
