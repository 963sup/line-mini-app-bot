package domain

import "errors"

var (
	ErrQualificationConflict = errors.New("membership relationship version changed")
	ErrInvalidQualification  = errors.New("invalid team membership change")
)

type TeamMembership struct {
	Qualification
	Version int64
}

func (m TeamMembership) Grant(role, parentKind, parentID string, parent Qualification, expectedVersion int64) (TeamMembership, error) {
	if m.Version != expectedVersion {
		return TeamMembership{}, ErrQualificationConflict
	}
	if (role != "member" && role != "maintainer") || (parentKind != "organization" && parentKind != "enterprise") ||
		parentID == "" || !parent.ActiveFor(parentKind) || !m.valid() {
		return TeamMembership{}, ErrInvalidQualification
	}
	current := m.ActiveFor("team") && m.MatchesParent(parentKind, parentID, parent)
	if current && m.Role == role {
		return m, nil
	}
	if !current {
		m.Generation++
	}
	m.Status, m.Role = "active", role
	m.ParentKind, m.ParentID, m.ParentGeneration = parentKind, parentID, parent.Generation
	m.Version++
	return m, nil
}

func (m TeamMembership) Revoke(expectedVersion int64) (TeamMembership, error) {
	if m.Version != expectedVersion {
		return TeamMembership{}, ErrQualificationConflict
	}
	if !m.valid() || m.Version == 0 {
		return TeamMembership{}, ErrInvalidQualification
	}
	if m.Status != "removed" {
		m.Status = "removed"
		m.Version++
	}
	return m, nil
}

func (m TeamMembership) valid() bool {
	if m.Version == 0 {
		return m.Qualification == (Qualification{})
	}
	return m.Version > 0 && m.Generation > 0 && m.Generation <= m.Version &&
		(m.Status == "active" || m.Status == "removed") &&
		(m.Role == "member" || m.Role == "maintainer")
}
