package entity

import "testing"

func TestLegacyOrganizationVisibility(t *testing.T) {
	for _, tc := range []struct {
		name, mode, actor        string
		owner, disabled, visible bool
	}{
		{name: "old owner", owner: true, actor: "owner", visible: true},
		{name: "explicit legacy member", mode: "legacy", owner: true, actor: "member", visible: true},
		{name: "unrelated actor", owner: true, actor: "outsider"},
		{name: "empty actor", owner: true},
		{name: "no owner", actor: "member"},
		{name: "disabled owner", owner: true, disabled: true, actor: "owner"},
		{name: "formal membership owner", owner: true, mode: "membership", actor: "owner"},
		{name: "unknown mode member", owner: true, mode: "future", actor: "member"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			org := Organization{MemberUserIDs: []string{"member", ""}, Disabled: tc.disabled, MembershipMode: tc.mode}
			if tc.owner {
				org.OwnerUserID = "owner"
			}
			if got := org.VisibleLegacy(tc.actor); got != tc.visible {
				t.Fatalf("visible = %v, want %v", got, tc.visible)
			}
		})
	}
}
