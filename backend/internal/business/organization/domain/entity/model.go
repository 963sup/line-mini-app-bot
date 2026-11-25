package entity

import "slices"

type Organization struct {
	Name           string
	ID             string
	EnterpriseID   string
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

func (o Organization) AttachEnterprise(actor, enterpriseID string) (Organization, error) {
	if !o.LegacyEnabled() || actor != o.OwnerUserID {
		return o, ErrNotFound
	}
	if enterpriseID == "" {
		return o, ErrInvalid
	}
	if o.EnterpriseID != "" && o.EnterpriseID != enterpriseID {
		return o, ErrConflict
	}
	o.EnterpriseID = enterpriseID
	return o, nil
}
