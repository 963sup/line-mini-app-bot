package organizationdb

import (
	organizationApplicationDto "github.com/963sup/line-mini-app-bot/internal/business/organization/application/dto"
	organizationApplicationQuery "github.com/963sup/line-mini-app-bot/internal/business/organization/application/query"
	organizationDomainEntity "github.com/963sup/line-mini-app-bot/internal/business/organization/domain/entity"
)

import (
	"context"
	"slices"
	"strings"

	"cloud.google.com/go/firestore"
	"github.com/963sup/line-mini-app-bot/internal/business/internal/adapter/firestore"
	membership "github.com/963sup/line-mini-app-bot/internal/business/membership/domain"

	"github.com/963sup/line-mini-app-bot/internal/platform/firestorekey"
	"github.com/google/uuid"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type Store struct {
	DB      *firestore.Client
	Channel string
	// EnterpriseNavigation reads facts using the current Firestore transaction.
	EnterpriseNavigation func(*firestore.Transaction, *firestore.DocumentRef, string) (map[string]string, bool, error)
	CanCreateEnterprise  func(*firestore.Transaction, *firestore.DocumentRef, string) (bool, error)
	CreateEnterprise     func(*firestore.Transaction, *firestore.DocumentRef, string, string, string) error
}

func (s Store) Context(ctx context.Context, user string, preference string, write bool) (*organizationApplicationDto.Context, error) {
	return orgstore.Run(ctx, s.DB, s.Channel, func(tx *firestore.Transaction, root *firestore.DocumentRef) (*organizationApplicationDto.Context, error) {
		if write && uuid.Validate(preference) != nil {
			return nil, organizationDomainEntity.ErrInvalid
		}
		orgs := root.Collection("organizations")
		pref := root.Collection("organizationPreferences").Doc(firestorekey.Digest([]byte(user)))
		docs, err := tx.Documents(orgs.WhereEntity(firestore.OrFilter{Filters: []firestore.EntityFilter{
			firestore.PropertyFilter{Path: "ownerUserId", Operator: "==", Value: user},
			firestore.PropertyFilter{Path: "memberUserIds", Operator: "array-contains", Value: user},
		}})).GetAll()
		if err != nil {
			return nil, err
		}
		rows := make([]organizationApplicationDto.VisibleOrganization, 0, len(docs))
		labels := map[string]string{}
		for _, d := range docs {
			var o orgstore.Document
			if err := d.DataTo(&o); err != nil {
				return nil, err
			}
			if !o.VisibleLegacyOrganization(user) {
				continue
			}
			if _, ok := labels[o.Enterprise]; !ok && o.Enterprise != "" {
				ent, err := orgstore.Read(tx, root.Collection("enterprises").Doc(o.Enterprise))
				if err != nil {
					return nil, err
				}
				labels[o.Enterprise] = ent.Name
			}
			rows = append(rows, organizationApplicationDto.VisibleOrganization{ID: d.Ref.ID, Name: o.Name, EnterpriseID: o.Enterprise, EnterpriseName: labels[o.Enterprise]})
		}
		selected := preference
		if !write {
			d, err := tx.Get(pref)
			if err == nil {
				value, err := d.DataAt("organizationId")
				if err != nil {
					return nil, err
				}
				selected, _ = value.(string)
			} else if status.Code(err) != codes.NotFound {
				return nil, err
			}
		}
		var enterprises map[string]string
		var canCreateEnterprise bool
		if s.EnterpriseNavigation != nil {
			var err error
			enterprises, canCreateEnterprise, err = s.EnterpriseNavigation(tx, root, user)
			if err != nil {
				return nil, err
			}
		}
		result := organizationApplicationQuery.BuildContextWithNavigation(rows, selected, enterprises, canCreateEnterprise)
		if write {
			if result.Selected == nil || result.Selected.ID != selected {
				return nil, organizationDomainEntity.ErrNotFound
			}
			return &result, tx.Set(pref, map[string]any{"organizationId": selected})
		}
		return &result, nil
	})
}

func (s Store) Create(ctx context.Context, user string, name string) (*organizationDomainEntity.Organization, error) {
	id := uuid.NewString()
	return orgstore.Run(ctx, s.DB, s.Channel, func(tx *firestore.Transaction, root *firestore.DocumentRef) (*organizationDomainEntity.Organization, error) {
		return createOrganization(tx, root, user, id, strings.TrimSpace(name))
	})
}

func createOrganization(tx *firestore.Transaction, root *firestore.DocumentRef, user, id, name string) (*organizationDomainEntity.Organization, error) {
	if name == "" || len([]rune(name)) > 120 {
		return nil, organizationDomainEntity.ErrInvalid
	}
	resource, err := membership.NewResourceRef(root.ID, "organization", id)
	if err != nil {
		return nil, organizationDomainEntity.ErrInvalid
	}
	opening, err := membership.OpenScope(resource, user)
	if err != nil {
		return nil, organizationDomainEntity.ErrInvalid
	}
	// Legacy mode stores the resource and its first owner in one document.
	org := &organizationDomainEntity.Organization{ID: opening.Resource().ID(), OwnerUserID: opening.OwnerID(), Name: name}
	return org, tx.Create(root.Collection("organizations").Doc(org.ID), map[string]any{
		"ownerUserId": org.OwnerUserID, "createdBy": user, "name": name,
	})
}

func (s Store) Transfer(ctx context.Context, user string, organizationId, target string) (*organizationDomainEntity.Organization, error) {
	return orgstore.Run(ctx, s.DB, s.Channel, func(tx *firestore.Transaction, root *firestore.DocumentRef) (*organizationDomainEntity.Organization, error) {
		if uuid.Validate(organizationId) != nil || len(organizationId) != 36 {
			return nil, organizationDomainEntity.ErrInvalid
		}
		ref := root.Collection("organizations").Doc(organizationId)
		org, e := orgstore.Read(tx, ref)
		if e != nil {
			return nil, e
		}
		if org.Owner != user || !org.LegacyOrganizationEnabled() {
			return nil, organizationDomainEntity.ErrNotFound
		}
		successor := target
		if !orgstore.UserID.MatchString(successor) {
			successor = ""
			for _, member := range org.Members {
				if orgstore.UserID.MatchString(member) && membership.MemberRef(organizationId, member) == target {
					successor = member
					break
				}
			}
		}
		current := organizationDomainEntity.Organization{ID: organizationId, Name: org.Name, OwnerUserID: org.Owner}
		result, err := current.TransferOwnership(user, successor, slices.Contains(org.Members, successor))
		if err != nil {
			return nil, err
		}
		if successor == user {
			return &result, nil
		}
		return &result, tx.Update(ref, []firestore.Update{
			{Path: "ownerUserId", Value: successor},
			{Path: "memberUserIds", Value: firestore.ArrayUnion(user)},
			{Path: "membershipOwnerVersion", Value: firestore.Increment(1)},
		})
	})
}
