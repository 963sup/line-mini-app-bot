package domain

// Qualification describes a stored relationship after its resource is checked.
type Qualification struct {
	Status, Role         string
	Generation           int64
	ParentKind, ParentID string
	ParentGeneration     int64
}

func (q Qualification) ActiveFor(kind string) bool {
	if q.Status != "active" || q.Generation <= 0 {
		return false
	}
	switch kind {
	case "organization":
		return q.Role == "member"
	case "enterprise":
		return q.Role == "member" || q.Role == "manager" || q.Role == "owner"
	case "team":
		return q.Role == "member" || q.Role == "maintainer"
	default:
		return false
	}
}

func (q Qualification) MatchesParent(kind, id string, parent Qualification) bool {
	return q.ParentKind == kind && q.ParentID == id &&
		parent.ActiveFor(kind) && q.ParentGeneration == parent.Generation
}
