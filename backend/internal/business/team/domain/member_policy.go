// Package domain owns Team management rules.
package domain

// CanManageMember governs both grant and revoke. The caller supplies authority
// verified against the current management scope and relationship generation.
func CanManageMember(parentManager, teamMaintainer bool, currentRole, requestedRole string) bool {
	if (currentRole != "" && currentRole != "member" && currentRole != "maintainer") ||
		(requestedRole != "" && requestedRole != "member" && requestedRole != "maintainer") {
		return false
	}
	if parentManager {
		return true
	}
	return teamMaintainer && currentRole != "maintainer" && requestedRole != "maintainer"
}
