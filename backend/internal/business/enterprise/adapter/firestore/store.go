package enterprisedb

import (
	"github.com/963sup/line-mini-app-bot/internal/business/enterprise/application/port"
	organizationDomainEntity "github.com/963sup/line-mini-app-bot/internal/business/organization/domain/entity"
)

import (
	"context"
	"errors"
	"slices"
	"strings"

	"github.com/963sup/line-mini-app-bot/internal/business/enterprise/domain/policy"

	"cloud.google.com/go/firestore"
	"github.com/963sup/line-mini-app-bot/internal/business/enterprise/application/dto"

	"github.com/963sup/line-mini-app-bot/internal/business/internal/adapter/firestore"
	groupport "github.com/963sup/line-mini-app-bot/internal/linegroup/application/port"
	groupentity "github.com/963sup/line-mini-app-bot/internal/linegroup/domain/entity"

	"github.com/963sup/line-mini-app-bot/internal/platform/firestore/key"
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
	OrganizationMember func(*firestore.Transaction, *firestore.DocumentRef, string, string) (bool, error)
	TeamMember         func(*firestore.Transaction, *firestore.DocumentRef, string, string, string, string) (bool, error)
}

type grant struct {
	Enabled      bool   `firestore:"enabled"`
	EnterpriseID string `firestore:"enterpriseId"`
}

func Provisioning(tx *firestore.Transaction, root *firestore.DocumentRef, user string) (grant, *firestore.DocumentRef, error) {
	ref := root.Collection("enterpriseEntitlements").Doc(firestorekey.Digest([]byte(user)))
	d, e := tx.Get(ref)
	var g grant
	if status.Code(e) == codes.NotFound {
		return g, ref, nil
	}
	if e == nil {
		e = d.DataTo(&g)
	}
	return g, ref, e
}

// CanStartCreation checks the same provisioning decision used by creation.
func CanStartCreation(tx *firestore.Transaction, root *firestore.DocumentRef, actor string) (bool, error) {
	g, _, err := Provisioning(tx, root, actor)
	return policy.CanProvision(g.Enabled, g.EnterpriseID, false), err
}

// CreateFromConversation joins the caller's conversation transaction.
func CreateFromConversation(tx *firestore.Transaction, root *firestore.DocumentRef, actor, id, name string) error {
	_, err := CreateInTransaction(tx, root, actor, id, name, false)
	return err
}

// NavigationFacts supplies Enterprise-owned visibility and provisioning facts
// from the transaction used by the organization navigation query.
func NavigationFacts(tx *firestore.Transaction, root *firestore.DocumentRef, user string) (map[string]string, bool, error) {
	canCreate, err := CanStartCreation(tx, root, user)
	if err != nil {
		return nil, false, err
	}
	labels := map[string]string{}
	// Keep the array-contains queries separate for Firestore compatibility.
	for _, q := range []firestore.Query{
		root.Collection("enterprises").Where("ownerUserId", "==", user),
		root.Collection("enterprises").Where("ownerUserIds", "array-contains", user),
		root.Collection("enterprises").Where("managerUserIds", "array-contains", user),
	} {
		docs, err := tx.Documents(q).GetAll()
		if err != nil {
			return nil, false, err
		}
		for _, doc := range docs {
			var ent orgstore.Document
			if err := doc.DataTo(&ent); err != nil {
				return nil, false, err
			}
			if legacyEnterpriseEnabled(ent) {
				labels[doc.Ref.ID] = ent.Name
			}
		}
	}
	return labels, canCreate, nil
}

func (s Store) CreateEnterprise(ctx context.Context, user, name string) (*dto.Link, error) {
	id := uuid.NewString()
	return orgstore.Run(ctx, s.DB, s.Channel, func(tx *firestore.Transaction, root *firestore.DocumentRef) (*dto.Link, error) {
		return CreateInTransaction(tx, root, user, id, strings.TrimSpace(name), true)
	})
}

func CreateInTransaction(tx *firestore.Transaction, root *firestore.DocumentRef, user, id, name string, allowExisting bool) (*dto.Link, error) {
	if name == "" || len([]rune(name)) > 120 {
		return nil, organizationDomainEntity.ErrInvalid
	}
	g, ref, e := Provisioning(tx, root, user)
	if e != nil {
		return nil, e
	}
	if !policy.CanProvision(g.Enabled, g.EnterpriseID, allowExisting) {
		return nil, organizationDomainEntity.ErrForbidden
	}
	if g.EnterpriseID != "" {
		existing, e := orgstore.Read(tx, root.Collection("enterprises").Doc(g.EnterpriseID))
		if e != nil {
			return nil, e
		}
		if !legacyEnterpriseEnabled(existing) || !existing.IsOwner(user) {
			return nil, organizationDomainEntity.ErrForbidden
		}
		return &dto.Link{ID: g.EnterpriseID, Name: existing.Name}, nil
	}
	if e := tx.Create(root.Collection("enterprises").Doc(id), map[string]any{"name": name, "ownerUserIds": []string{user}, "createdBy": user}); e != nil {
		return nil, e
	}
	return &dto.Link{ID: id, Name: name}, tx.Update(ref, []firestore.Update{{Path: "enterpriseId", Value: id}})
}

func legacyEnterpriseEnabled(ent orgstore.Document) bool {
	return !ent.Disabled && (ent.MembershipMode == "" || ent.MembershipMode == "legacy")
}

func (s Store) Enterprise(ctx context.Context, user, id string) (*dto.Detail, error) {
	if uuid.Validate(id) != nil {
		return nil, organizationDomainEntity.ErrInvalid
	}
	return orgstore.Run(ctx, s.DB, s.Channel, func(tx *firestore.Transaction, root *firestore.DocumentRef) (*dto.Detail, error) {
		ent, e := orgstore.Read(tx, root.Collection("enterprises").Doc(id))
		if e != nil {
			return nil, e
		}
		if ent.Disabled || (ent.MembershipMode != "" && ent.MembershipMode != "legacy" && ent.MembershipMode != "membership") {
			return nil, organizationDomainEntity.ErrNotFound
		}
		legacy := legacyEnterpriseEnabled(ent)
		owner := legacy && ent.IsOwner(user)
		admin := owner || (legacy && slices.Contains(ent.Managers, user))
		view := dto.Detail{ID: id, Name: ent.Name, CanManage: owner, CanManageTeams: admin, Organizations: []dto.Link{}, Teams: []dto.Link{}}
		docs, e := tx.Documents(root.Collection("organizations").Where("enterpriseId", "==", id)).GetAll()
		if e != nil {
			return nil, e
		}
		for _, d := range docs {
			var org orgstore.Document
			if e := d.DataTo(&org); e != nil {
				return nil, e
			}
			visible := admin || org.VisibleLegacyOrganization(user)
			if !visible && org.MembershipMode == "membership" && !org.Disabled && s.OrganizationMember != nil {
				visible, e = s.OrganizationMember(tx, root, d.Ref.ID, user)
				if e != nil {
					return nil, e
				}
			}
			if view.CanManageTeams && org.LegacyOrganizationEnabled() {
				view.AssignableOrganizations = append(view.AssignableOrganizations, dto.Link{ID: d.Ref.ID, Name: org.Name})
			}
			if visible {
				view.Organizations = append(view.Organizations, dto.Link{ID: d.Ref.ID, Name: org.Name})
			}
		}
		if !admin && len(view.Organizations) == 0 {
			return nil, organizationDomainEntity.ErrNotFound
		}
		assigned := map[string]bool{}
		if !admin {
			for _, org := range view.Organizations {
				links, err := tx.Documents(root.Collection("organizations").Doc(org.ID).Collection("enterpriseTeams").Query).GetAll()
				if err != nil {
					return nil, err
				}
				for _, link := range links {
					assigned[link.Ref.ID] = true
				}
			}
		}
		docs, e = tx.Documents(root.Collection("teams").Where("managementId", "==", id)).GetAll()
		if e != nil {
			return nil, e
		}
		for _, d := range docs {
			var team orgstore.Document
			if e := d.DataTo(&team); e != nil {
				return nil, e
			}
			if team.Scope != "enterprise" || team.Disabled {
				continue
			}
			switch team.MembershipMode {
			case "membership":
				if (!admin && !assigned[d.Ref.ID]) || team.Status != "ACTIVE" || s.TeamMember == nil {
					continue
				}
				allowed, err := s.TeamMember(tx, root, d.Ref.ID, team.Scope, team.Manager, user)
				if err != nil {
					return nil, err
				}
				if allowed {
					view.Teams = append(view.Teams, dto.Link{ID: d.Ref.ID, Name: team.Name})
				}
				continue
			case "", "legacy":
			default:
				continue
			}
			if !admin {
				if !assigned[d.Ref.ID] {
					continue
				}
				if s.Groups == nil {
					return nil, groupport.ErrUnavailable
				}
				_, e := s.Groups.Get(ctx, user, d.Ref.ID)
				if e != nil {
					if errors.Is(e, groupentity.ErrNotFound) {
						continue
					}
					return nil, e
				}
			}
			view.Teams = append(view.Teams, dto.Link{ID: d.Ref.ID, Name: team.Name})
		}
		return &view, nil
	})
}

func (s Store) RunOwnerChange(ctx context.Context, id, owner string, apply func(port.OwnerChange) (*dto.Link, error)) (*dto.Link, error) {
	if uuid.Validate(id) != nil || !orgstore.UserID.MatchString(owner) {
		return nil, organizationDomainEntity.ErrInvalid
	}
	return orgstore.Run(ctx, s.DB, s.Channel, func(tx *firestore.Transaction, root *firestore.DocumentRef) (*dto.Link, error) {
		return apply(ownerChange{enterpriseWork{tx: tx, ref: root.Collection("enterprises").Doc(id)}, owner})
	})
}

func (s Store) RunOrganizationLink(ctx context.Context, id, organization string, apply func(port.OrganizationLink) (*dto.Link, error)) (*dto.Link, error) {
	if uuid.Validate(id) != nil || uuid.Validate(organization) != nil {
		return nil, organizationDomainEntity.ErrInvalid
	}
	return orgstore.Run(ctx, s.DB, s.Channel, func(tx *firestore.Transaction, root *firestore.DocumentRef) (*dto.Link, error) {
		return apply(organizationLink{enterpriseWork{tx: tx, ref: root.Collection("enterprises").Doc(id)}, root.Collection("organizations").Doc(organization)})
	})
}

type enterpriseWork struct {
	tx  *firestore.Transaction
	ref *firestore.DocumentRef
}

func (w enterpriseWork) Enterprise(actor string) (port.EnterpriseFacts, error) {
	ent, err := orgstore.Read(w.tx, w.ref)
	if err != nil {
		return port.EnterpriseFacts{}, err
	}
	return port.EnterpriseFacts{Name: ent.Name, Active: legacyEnterpriseEnabled(ent), ActorIsOwner: ent.IsOwner(actor)}, nil
}

type ownerChange struct {
	enterpriseWork
	owner string
}

func (w ownerChange) AddOwner() error {
	return w.tx.Update(w.ref, []firestore.Update{{Path: "ownerUserIds", Value: firestore.ArrayUnion(w.owner)}})
}

type organizationLink struct {
	enterpriseWork
	organization *firestore.DocumentRef
}

func (w organizationLink) Organization(actor string) (port.OrganizationFacts, error) {
	org, err := orgstore.Read(w.tx, w.organization)
	if err != nil {
		return port.OrganizationFacts{}, err
	}
	return port.OrganizationFacts{Active: org.LegacyOrganizationEnabled(), ActorIsOwner: org.Owner == actor, EnterpriseID: org.Enterprise}, nil
}

func (w organizationLink) Attach() error {
	return w.tx.Update(w.organization, []firestore.Update{{Path: "enterpriseId", Value: w.ref.ID}})
}
