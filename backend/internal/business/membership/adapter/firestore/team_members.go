package membershipdb

import (
	"context"
	"github.com/963sup/line-mini-app-bot/internal/platform/line/lineid"

	"cloud.google.com/go/firestore"
	"github.com/963sup/line-mini-app-bot/internal/business/membership/application"
	"github.com/963sup/line-mini-app-bot/internal/business/membership/domain"
	"github.com/963sup/line-mini-app-bot/internal/platform/firestore/key"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (s Store) ChangeTeamMember(ctx context.Context, cmd application.TeamMemberChange, decide func(application.TeamMemberFacts, bool) (domain.TeamMembership, error)) error {
	if !lineid.ValidGroup(cmd.TeamID) {
		return domain.ErrInvalidQualification
	}
	root := s.DB.Collection("channels").Doc(s.Channel)
	return s.DB.RunTransaction(ctx, func(_ context.Context, tx *firestore.Transaction) error {
		facts, err := teamMemberFacts(tx, root, cmd)
		if err != nil {
			return err
		}
		receipt := root.Collection("membershipAudits").Doc(firestorekey.Encode("team_member_change", cmd.TeamID, cmd.ActorID, cmd.RequestID))
		fingerprint := firestorekey.JSONDigest(cmd)
		prior, err := tx.Get(receipt)
		replay := err == nil
		if replay {
			value, err := prior.DataAt("fingerprint")
			if err != nil {
				return err
			}
			if value != fingerprint {
				return domain.ErrQualificationConflict
			}
		} else if status.Code(err) != codes.NotFound {
			return err
		}

		next, err := decide(facts, replay)
		if err != nil || replay {
			return err
		}
		if next != facts.Target {
			ref := relationshipRef(root, "team", cmd.TeamID, cmd.TargetID)
			record := teamMemberRecord(cmd.TeamID, cmd.TargetID, next)
			if err := tx.Set(ref, record, firestore.MergeAll); err != nil {
				return err
			}
		}
		return tx.Create(receipt, map[string]any{
			"resourceType": "team", "resourceId": cmd.TeamID, "actorUserId": cmd.ActorID, "targetUserId": cmd.TargetID,
			"requestId": cmd.RequestID, "fingerprint": fingerprint, "action": "team_member_change",
			"beforeVersion": facts.Target.Version, "afterVersion": next.Version, "role": next.Role, "status": next.Status,
			"at": firestore.ServerTimestamp,
		})
	})
}

func teamMemberFacts(tx *firestore.Transaction, root *firestore.DocumentRef, cmd application.TeamMemberChange) (application.TeamMemberFacts, error) {
	var facts application.TeamMemberFacts
	doc, err := tx.Get(root.Collection("teams").Doc(cmd.TeamID))
	if status.Code(err) == codes.NotFound {
		return facts, application.ErrForbidden
	}
	if err != nil {
		return facts, err
	}
	var team struct {
		Scope    string `firestore:"managementScope"`
		Parent   string `firestore:"managementId"`
		Mode     string `firestore:"membershipMode"`
		Status   string `firestore:"status"`
		Disabled bool   `firestore:"disabled"`
	}
	if err := doc.DataTo(&team); err != nil {
		return facts, err
	}
	if team.Mode != "membership" || team.Status != "ACTIVE" || team.Disabled {
		return facts, application.ErrForbidden
	}
	parent, err := parentQualification(tx, root, team.Scope, team.Parent, cmd.ActorID)
	if err != nil {
		return facts, err
	}
	if !parent.ActiveFor(team.Scope) {
		return facts, application.ErrForbidden
	}
	facts.ParentKind, facts.ParentID = team.Scope, team.Parent
	facts.ParentManager, err = managesParent(tx, root, team.Scope, team.Parent, cmd.ActorID, parent)
	if err != nil {
		return facts, err
	}
	actor, err := readQualification(tx, root, "team", cmd.TeamID, cmd.ActorID)
	if err != nil {
		return facts, err
	}
	facts.TeamMaintainer = actor.Role == "maintainer" && actor.ActiveFor("team") && actor.MatchesParent(team.Scope, team.Parent, parent)
	facts.Target, err = readRelationship(tx, root, "team", cmd.TeamID, cmd.TargetID)
	if err == nil && cmd.Role != "" {
		facts.TargetParent, err = parentQualification(tx, root, team.Scope, team.Parent, cmd.TargetID)
	}
	return facts, err
}

// OpenTeamMembership prepares the first maintainer inside Team's activation
// transaction. Team remains responsible for binding the resource.
func OpenTeamMembership(tx *firestore.Transaction, root *firestore.DocumentRef, teamID, kind, parentID, actor, maintainer string, replay bool) error {
	parent, err := parentQualification(tx, root, kind, parentID, actor)
	if err != nil {
		return err
	}
	allowed, err := managesParent(tx, root, kind, parentID, actor, parent)
	if err != nil {
		return err
	}
	if !allowed {
		return application.ErrForbidden
	}
	if replay {
		return nil
	}
	target := parent
	if maintainer != actor {
		target, err = parentQualification(tx, root, kind, parentID, maintainer)
		if err != nil {
			return err
		}
	}
	if !target.ActiveFor(kind) {
		return application.ErrForbidden
	}
	next, err := (domain.TeamMembership{}).Grant("maintainer", kind, parentID, target, 0)
	if err != nil {
		return err
	}
	if err := tx.Create(relationshipRef(root, "team", teamID, maintainer), teamMemberRecord(teamID, maintainer, next)); err != nil {
		return err
	}
	return tx.Create(root.Collection("membershipAudits").Doc(firestorekey.Encode("team_opened", teamID)), map[string]any{
		"resourceType": "team", "resourceId": teamID, "actorUserId": actor,
		"targetUserId": maintainer, "action": "team_opened", "at": firestore.ServerTimestamp,
	})
}

func managesParent(tx *firestore.Transaction, root *firestore.DocumentRef, kind, parentID, actor string, parent domain.Qualification) (bool, error) {
	if !parent.ActiveFor(kind) {
		return false, nil
	}
	if kind == "enterprise" {
		return parent.Role == "owner" || parent.Role == "manager", nil
	}
	if kind != "organization" {
		return false, nil
	}
	state, err := tx.Get(root.Collection("resourceMembershipStates").Doc(firestorekey.Encode(kind, parentID)))
	if status.Code(err) == codes.NotFound {
		return false, nil
	}
	if err != nil {
		return false, err
	}
	var scope struct {
		Owner string `firestore:"ownerUserId"`
	}
	if err := state.DataTo(&scope); err != nil {
		return false, err
	}
	return scope.Owner == actor, nil
}

func teamMemberRecord(teamID, user string, next domain.TeamMembership) map[string]any {
	return map[string]any{
		"resourceType": "team", "resourceId": teamID, "userId": user,
		"status": next.Status, "role": next.Role, "generation": next.Generation, "version": next.Version,
		"parentKind": next.ParentKind, "parentId": next.ParentID, "parentGeneration": next.ParentGeneration,
		"updatedAt": firestore.ServerTimestamp,
	}
}
