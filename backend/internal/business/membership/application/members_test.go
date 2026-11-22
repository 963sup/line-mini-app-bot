package application

import "testing"

type rosterWork struct{ Work }

func (rosterWork) Roster(string) (Roster, error) {
	return Roster{OwnerID: "owner", UserIDs: []string{"owner", "member"}}, nil
}

func TestRosterActions(t *testing.T) {
	service := Service{Store: invitationRetryStore(func(fn func(Work) error) error { return fn(rosterWork{}) })}
	for _, actor := range []string{"owner", "member"} {
		roster, err := service.ListMembers(t.Context(), actor, "org")
		owner := actor == "owner"
		if err != nil || len(roster.People) != 2 || roster.IsOwner != owner || roster.CanTransferOwnership != owner {
			t.Fatalf("unexpected roster for %s: %+v, %v", actor, roster, err)
		}
		for _, member := range roster.People {
			allowed := owner && !member.Owner
			if member.CanRemove != allowed || member.CanReceiveOwnership != allowed {
				t.Errorf("wrong target actions for %s: %+v", actor, member)
			}
		}
	}
}
