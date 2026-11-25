package teamdb

import (
	teamapp "github.com/963sup/line-mini-app-bot/internal/business/team/application"
	teamApplicationDto "github.com/963sup/line-mini-app-bot/internal/business/team/application/dto"
	teamdomain "github.com/963sup/line-mini-app-bot/internal/business/team/domain"
)

import (
	membershipapp "github.com/963sup/line-mini-app-bot/internal/business/membership/application"
	organizationDomainEntity "github.com/963sup/line-mini-app-bot/internal/business/organization/domain/entity"
)

import (
	"context"
	"errors"
	"slices"
	"strings"
	"sync"

	"cloud.google.com/go/firestore"

	"github.com/963sup/line-mini-app-bot/internal/business/internal/adapter/firestore"
	groupport "github.com/963sup/line-mini-app-bot/internal/business/team/application/port"
	groupentity "github.com/963sup/line-mini-app-bot/internal/business/team/domain/entity"
	"github.com/963sup/line-mini-app-bot/internal/platform/line/lineid"

	"github.com/google/uuid"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type Store struct {
	DB      *firestore.Client
	Channel string
	Groups  interface {
		Get(context.Context, string, string) (groupentity.Group, error)
	}
	LINEActivationMember groupport.Membership
	OrganizationMember   func(*firestore.Transaction, *firestore.DocumentRef, string, string) (bool, error)
	TeamMember           func(*firestore.Transaction, *firestore.DocumentRef, string, string, string, string) (bool, error)
	OpenMembership       func(*firestore.Transaction, *firestore.DocumentRef, string, string, string, string, string, bool) error
}

// ApplyGroupObservation keeps the existing Team lifecycle linked to the accepted
// LINE group state. It preserves Team management and membership fields.
func ApplyGroupObservation(tx *firestore.Transaction, channel *firestore.DocumentRef, teamID, name, status string) error {
	return tx.Set(channel.Collection("teams").Doc(teamID), map[string]any{"name": name, "status": status}, firestore.MergeAll)
}

// IsTeamMember authorizes only an active team with formal membership.
func (s Store) IsTeamMember(ctx context.Context, teamID, actor string) (bool, error) {
	if s.TeamMember == nil {
		return false, groupport.ErrUnavailable
	}
	result, err := orgstore.Run(ctx, s.DB, s.Channel, func(tx *firestore.Transaction, root *firestore.DocumentRef) (*bool, error) {
		team, err := orgstore.Read(tx, root.Collection("teams").Doc(teamID))
		if status.Code(err) == codes.NotFound {
			return nil, nil
		}
		if err != nil {
			return nil, err
		}
		if team.Disabled || team.Status != "ACTIVE" || team.MembershipMode != "membership" {
			return nil, nil
		}
		ok, err := s.TeamMember(tx, root, teamID, team.Scope, team.Manager, actor)
		return &ok, err
	})
	return err == nil && result != nil && *result, err
}

func assignmentAuthority(target orgstore.Document, owner bool, user string) teamdomain.AssignmentAuthority {
	return teamdomain.AssignmentAuthority{
		Disabled: target.Disabled, MembershipMode: target.MembershipMode,
		Owner: owner, Manager: slices.Contains(target.Managers, user),
	}
}

func (s Store) Manage(ctx context.Context, user string, teamId, scope, managementId string) error {
	_, err := orgstore.Run(ctx, s.DB, s.Channel, func(tx *firestore.Transaction, root *firestore.DocumentRef) (*struct{}, error) {
		if !lineid.ValidGroup(teamId) || uuid.Validate(managementId) != nil || (scope != "organization" && scope != "enterprise") {
			return nil, organizationDomainEntity.ErrInvalid
		}
		target, e := orgstore.Read(tx, root.Collection(scope+"s").Doc(managementId))
		if e != nil {
			return nil, e
		}
		if scope == "organization" && !target.LegacyOrganizationEnabled() {
			return nil, organizationDomainEntity.ErrNotFound
		}
		owner := target.Owner == user
		if scope == "enterprise" {
			owner = target.IsOwner(user)
		}
		if !assignmentAuthority(target, owner, user).AllowsAssignment() {
			return nil, organizationDomainEntity.ErrNotFound
		}
		ref := root.Collection("teams").Doc(teamId)
		team, e := orgstore.Read(tx, ref)
		if e != nil {
			return nil, e
		}
		if team.Disabled || team.Status != "ACTIVE" || (team.MembershipMode != "" && team.MembershipMode != "legacy") || !slices.Contains(team.Managers, user) {
			return nil, organizationDomainEntity.ErrNotFound
		}
		if team.Scope != "" && (team.Scope != scope || team.Manager != managementId) {
			return nil, organizationDomainEntity.ErrConflict
		}
		return nil, tx.Update(ref, []firestore.Update{{Path: "managementScope", Value: scope}, {Path: "managementId", Value: managementId}})
	})
	return err
}

// Activate opens an observed, unbound Team with independent Membership.
func (s Store) Activate(ctx context.Context, actor, teamID, scope, managementID, maintainer string) error {
	if !lineid.ValidGroup(teamID) || len(managementID) != 36 || uuid.Validate(managementID) != nil ||
		(scope != "organization" && scope != "enterprise") || !lineid.ValidUser(maintainer) {
		return organizationDomainEntity.ErrInvalid
	}
	if s.OpenMembership == nil {
		return groupport.ErrUnavailable
	}
	checkLINE := sync.OnceValues(func() (bool, error) {
		if s.LINEActivationMember == nil {
			return false, groupport.ErrUnavailable
		}
		return s.LINEActivationMember.IsMember(ctx, teamID, actor)
	})
	_, err := orgstore.Run(ctx, s.DB, s.Channel, func(tx *firestore.Transaction, root *firestore.DocumentRef) (*struct{}, error) {
		ref := root.Collection("teams").Doc(teamID)
		team, err := orgstore.Read(tx, ref)
		if err != nil {
			return nil, err
		}
		if team.Disabled || team.Status != "ACTIVE" {
			return nil, organizationDomainEntity.ErrNotFound
		}
		replay := team.MembershipMode == "membership" && team.Scope == scope &&
			team.Manager == managementID && team.InitialMaintainer == maintainer
		unbound := team.Scope == "" && team.Manager == "" &&
			(team.MembershipMode == "" || team.MembershipMode == "legacy")
		if !slices.Contains(team.Managers, actor) {
			if len(team.Managers) != 0 || (!unbound && !replay) {
				return nil, organizationDomainEntity.ErrNotFound
			}
			if !replay {
				allowed, err := checkLINE()
				if err != nil {
					return nil, err
				}
				if !allowed {
					return nil, organizationDomainEntity.ErrNotFound
				}
			}
		}
		if !unbound && !replay {
			return nil, organizationDomainEntity.ErrConflict
		}
		if err := s.OpenMembership(tx, root, teamID, scope, managementID, actor, maintainer, replay); err != nil {
			if errors.Is(err, membershipapp.ErrForbidden) {
				return nil, organizationDomainEntity.ErrNotFound
			}
			return nil, err
		}
		if replay {
			return nil, nil
		}
		return nil, tx.Update(ref, []firestore.Update{
			{Path: "managementScope", Value: scope},
			{Path: "managementId", Value: managementID},
			{Path: "membershipMode", Value: "membership"},
			{Path: "initialMaintainerUserId", Value: maintainer},
		})
	})
	if status.Code(err) == codes.AlreadyExists {
		return organizationDomainEntity.ErrConflict
	}
	return err
}

func (s Store) RunAssignment(ctx context.Context, organizationID, teamID string, apply func(teamapp.AssignmentWork) error) error {
	_, err := orgstore.Run(ctx, s.DB, s.Channel, func(tx *firestore.Transaction, root *firestore.DocumentRef) (*struct{}, error) {
		if uuid.Validate(organizationID) != nil || !lineid.ValidGroup(teamID) {
			return nil, organizationDomainEntity.ErrInvalid
		}
		return nil, apply(assignmentWork{tx: tx, root: root, organization: root.Collection("organizations").Doc(organizationID), teamID: teamID})
	})
	return err
}

type assignmentWork struct {
	tx                 *firestore.Transaction
	root, organization *firestore.DocumentRef
	teamID             string
}

func (w assignmentWork) Scope() (teamdomain.AssignmentScope, error) {
	org, err := orgstore.Read(w.tx, w.organization)
	if err != nil {
		return teamdomain.AssignmentScope{}, err
	}
	team, err := orgstore.Read(w.tx, w.root.Collection("teams").Doc(w.teamID))
	if err != nil {
		return teamdomain.AssignmentScope{}, err
	}
	return teamdomain.AssignmentScope{OrganizationActive: org.LegacyOrganizationEnabled(), TeamActive: team.Status == "ACTIVE", TeamDisabled: team.Disabled, EnterpriseID: org.Enterprise, TeamScope: team.Scope, TeamManagerID: team.Manager}, nil
}

func (w assignmentWork) Authority(enterpriseID, actor string) (teamdomain.AssignmentAuthority, error) {
	ent, err := orgstore.Read(w.tx, w.root.Collection("enterprises").Doc(enterpriseID))
	if err != nil {
		return teamdomain.AssignmentAuthority{}, err
	}
	return assignmentAuthority(ent, ent.IsOwner(actor), actor), nil
}

func (w assignmentWork) Save(assigned bool) error {
	link := w.organization.Collection("enterpriseTeams").Doc(w.teamID)
	if assigned {
		return w.tx.Set(link, map[string]any{"teamId": w.teamID})
	}
	return w.tx.Delete(link)
}

func (s Store) List(ctx context.Context, user string, organizationId string) (*teamApplicationDto.Listing, error) {
	type listing struct {
		result  teamApplicationDto.Listing
		pending []teamApplicationDto.Option
	}
	snapshot, err := orgstore.Run(ctx, s.DB, s.Channel, func(tx *firestore.Transaction, root *firestore.DocumentRef) (*listing, error) {
		if uuid.Validate(organizationId) != nil {
			return nil, organizationDomainEntity.ErrInvalid
		}
		ref := root.Collection("organizations").Doc(organizationId)
		org, e := orgstore.Read(tx, ref)
		if e != nil {
			return nil, e
		}
		visible := org.VisibleLegacyOrganization(user)
		if org.MembershipMode == "membership" && !org.Disabled && s.OrganizationMember != nil {
			visible, e = s.OrganizationMember(tx, root, organizationId, user)
			if e != nil {
				return nil, e
			}
		}
		if !visible {
			return nil, organizationDomainEntity.ErrNotFound
		}
		teams := root.Collection("teams")
		local, e := tx.Documents(teams.Where("managementId", "==", organizationId)).GetAll()
		if e != nil {
			return nil, e
		}
		candidates := map[string]orgstore.Document{}
		for _, doc := range local {
			var team orgstore.Document
			if e := doc.DataTo(&team); e != nil {
				return nil, e
			}
			if team.Scope == "organization" {
				candidates[doc.Ref.ID] = team
			}
		}
		links, e := tx.Documents(ref.Collection("enterpriseTeams").Query).GetAll()
		if e != nil {
			return nil, e
		}
		for _, doc := range links {
			team, e := orgstore.Read(tx, teams.Doc(doc.Ref.ID))
			if status.Code(e) == codes.NotFound {
				continue
			}
			if e != nil {
				return nil, e
			}
			scope := teamdomain.AssignmentScope{EnterpriseID: org.Enterprise, TeamScope: team.Scope, TeamManagerID: team.Manager}
			if scope.TeamManagedByEnterprise() {
				candidates[doc.Ref.ID] = team
			}
		}
		result := teamApplicationDto.Listing{Teams: []teamApplicationDto.Option{}, CanBindTeam: org.LegacyOrganizationEnabled() && assignmentAuthority(org, org.Owner == user, user).AllowsAssignment()}
		var pending []teamApplicationDto.Option
		for id, team := range candidates {
			switch team.MembershipMode {
			case "membership":
				if team.Disabled || team.Status != "ACTIVE" || s.TeamMember == nil {
					continue
				}
				allowed, err := s.TeamMember(tx, root, id, team.Scope, team.Manager, user)
				if err != nil {
					return nil, err
				}
				if allowed {
					result.Teams = append(result.Teams, teamApplicationDto.Option{ID: id, Name: team.Name, ManagementScope: team.Scope})
				}
				continue
			case "", "legacy":
				if team.Disabled {
					continue
				}
			default:
				continue
			}
			pending = append(pending, teamApplicationDto.Option{ID: id, ManagementScope: team.Scope})
		}
		return &listing{result, pending}, nil
	})
	if err != nil {
		return nil, err
	}
	result := snapshot.result
	for _, candidate := range snapshot.pending {
		if s.Groups == nil {
			return nil, groupport.ErrUnavailable
		}
		group, err := s.Groups.Get(ctx, user, candidate.ID)
		if errors.Is(err, groupentity.ErrNotFound) {
			continue
		}
		if err != nil {
			return nil, err
		}
		candidate.Name = group.Name
		result.Teams = append(result.Teams, candidate)
	}
	slices.SortFunc(result.Teams, func(a, b teamApplicationDto.Option) int { return strings.Compare(a.ID, b.ID) })
	return &result, nil
}
