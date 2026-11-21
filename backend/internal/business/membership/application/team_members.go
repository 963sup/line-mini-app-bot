package application

import (
	"context"

	"github.com/963sup/line-mini-app-bot/internal/business/membership/domain"
)

// TeamMemberStore binds all facts and the resulting write to one transaction.
type TeamMemberStore interface {
	ChangeTeamMember(context.Context, TeamMemberChange, func(TeamMemberFacts, bool) (domain.TeamMembership, error)) error
}

type TeamMemberChange struct {
	TeamID, ActorID, TargetID, RequestID string
	Role                                 string // Empty for revocation.
	ExpectedVersion                      int64
}

type TeamMemberFacts struct {
	ParentKind, ParentID          string
	ParentManager, TeamMaintainer bool
	TargetParent                  domain.Qualification
	Target                        domain.TeamMembership
}

type TeamMembers struct {
	Store TeamMemberStore
	// CanManage is supplied by Team policy at composition.
	CanManage func(parentManager, teamMaintainer bool, currentRole, requestedRole string) bool
}

func (s TeamMembers) Change(ctx context.Context, command TeamMemberChange) error {
	if command.TeamID == "" || command.ActorID == "" || command.TargetID == "" ||
		command.RequestID == "" || command.ExpectedVersion < 0 {
		return domain.ErrInvalidQualification
	}
	return s.Store.ChangeTeamMember(ctx, command, func(facts TeamMemberFacts, replay bool) (domain.TeamMembership, error) {
		if !s.CanManage(facts.ParentManager, facts.TeamMaintainer, facts.Target.Role, command.Role) {
			return domain.TeamMembership{}, ErrForbidden
		}
		if replay {
			return facts.Target, nil
		}
		if command.Role == "" {
			return facts.Target.Revoke(command.ExpectedVersion)
		}
		return facts.Target.Grant(command.Role, facts.ParentKind, facts.ParentID, facts.TargetParent, command.ExpectedVersion)
	})
}
