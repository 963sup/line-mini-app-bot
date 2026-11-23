package dto

// VisibleOrganization is supplied after membership and owner authorization.
type VisibleOrganization struct {
	ID, Name, EnterpriseID, EnterpriseName string
}

type Option struct {
	ID     string
	Name   string
	Detail *string
}

type Group struct {
	EnterpriseID  *string
	Label         string
	Organizations []Option
}

type Selected struct {
	ID         string
	Name       string
	GroupLabel string
}

type EnterpriseLink struct {
	ID   string
	Name string
}

type Context struct {
	Enterprises         []EnterpriseLink
	CanCreateEnterprise bool
	Groups              []Group
	Selected            *Selected
	CanSwitch           bool
}
