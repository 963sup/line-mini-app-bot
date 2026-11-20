package dto

// VisibleOrganization is supplied after membership and owner authorization.
type VisibleOrganization struct {
	ID, Name, EnterpriseID, EnterpriseName string
}

type Option struct {
	ID     string  `json:"organizationId"`
	Name   string  `json:"name"`
	Detail *string `json:"detail"`
}

type Group struct {
	EnterpriseID  *string  `json:"enterpriseId"`
	Label         string   `json:"label"`
	Organizations []Option `json:"organizations"`
}

type Selected struct {
	ID         string `json:"organizationId"`
	Name       string `json:"name"`
	GroupLabel string `json:"groupLabel"`
}

type EnterpriseLink struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

type Context struct {
	Enterprises         []EnterpriseLink `json:"enterprises"`
	CanCreateEnterprise bool             `json:"canCreateEnterprise"`
	Groups              []Group          `json:"groups"`
	Selected            *Selected        `json:"selected"`
	CanSwitch           bool             `json:"canSwitch"`
}
