package domain

type AssignmentScope struct {
	OrganizationActive bool
	TeamActive         bool
	TeamDisabled       bool
	EnterpriseID       string
	TeamScope          string
	TeamManagerID      string
}

func (s AssignmentScope) AllowsAssignment() bool {
	return s.OrganizationActive && s.TeamActive && !s.TeamDisabled && s.TeamManagedByEnterprise()
}

func (s AssignmentScope) TeamManagedByEnterprise() bool {
	return s.EnterpriseID != "" && s.TeamScope == "enterprise" && s.TeamManagerID == s.EnterpriseID
}

type AssignmentAuthority struct {
	Disabled       bool
	MembershipMode string
	Owner          bool
	Manager        bool
}

func (a AssignmentAuthority) AllowsAssignment() bool {
	return !a.Disabled && (a.MembershipMode == "" || a.MembershipMode == "legacy") &&
		(a.Owner || a.Manager)
}
