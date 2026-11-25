package organizationdb

import (
	organizationapp "github.com/963sup/line-mini-app-bot/internal/business/organization/application"
	organizationApplicationDto "github.com/963sup/line-mini-app-bot/internal/business/organization/application/dto"
	organizationApplicationQuery "github.com/963sup/line-mini-app-bot/internal/business/organization/application/query"
	organizationDomainEntity "github.com/963sup/line-mini-app-bot/internal/business/organization/domain/entity"
)

import (
	"context"
	"errors"
	"strings"

	"cloud.google.com/go/firestore"
	"github.com/963sup/line-mini-app-bot/internal/business/internal/adapter/firestore"
	membership "github.com/963sup/line-mini-app-bot/internal/business/membership/domain"

	"github.com/963sup/line-mini-app-bot/internal/platform/firestore/key"
	"github.com/google/uuid"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type Store struct {
	OrganizationMember func(*firestore.Transaction, *firestore.DocumentRef, string, string) (bool, error)
	DB                 *firestore.Client
	Channel            string
	// EnterpriseNavigation reads facts using the current Firestore transaction.
	EnterpriseNavigation func(*firestore.Transaction, *firestore.DocumentRef, string) (map[string]string, bool, error)
	CanCreateEnterprise  func(*firestore.Transaction, *firestore.DocumentRef, string) (bool, error)
	CreateEnterprise     func(*firestore.Transaction, *firestore.DocumentRef, string, string, string) error
}

// CanAccessOwner exposes only the organization membership decision required by
// another bounded context; repository ownership rules remain in Repository.
func (s Store) CanAccessOwner(ctx context.Context, ownerType, ownerID, actor string) (bool, error) {
	if ownerType != "ORGANIZATION" || uuid.Validate(ownerID) != nil || actor == "" {
		return false, nil
	}
	result, err := orgstore.Run(ctx, s.DB, s.Channel, func(tx *firestore.Transaction, root *firestore.DocumentRef) (*bool, error) {
		doc, err := orgstore.Read(tx, root.Collection("organizations").Doc(ownerID))
		if err != nil {
			return nil, err
		}
		allowed := doc.VisibleLegacyOrganization(actor)
		return &allowed, nil
	})
	if errors.Is(err, organizationDomainEntity.ErrNotFound) {
		return false, nil
	}
	return result != nil && *result, err
}

// ResolveRepositoryOwner returns the actor's explicit, still-visible
// organization preference; absence falls back to personal ownership.
func (s Store) ResolveRepositoryOwner(ctx context.Context, actor string) (string, string, error) {
	if actor == "" {
		return "", "", organizationDomainEntity.ErrInvalid
	}
	selected, err := orgstore.Run(ctx, s.DB, s.Channel, func(tx *firestore.Transaction, root *firestore.DocumentRef) (*string, error) {
		pref, err := tx.Get(root.Collection("organizationPreferences").Doc(firestorekey.Digest([]byte(actor))))
		if status.Code(err) == codes.NotFound {
			value := ""
			return &value, nil
		}
		if err != nil {
			return nil, err
		}
		value, err := pref.DataAt("organizationId")
		if err != nil {
			return nil, err
		}
		organizationID, _ := value.(string)
		if uuid.Validate(organizationID) != nil {
			return nil, organizationDomainEntity.ErrInvalid
		}
		doc, err := orgstore.Read(tx, root.Collection("organizations").Doc(organizationID))
		if err != nil {
			return nil, err
		}
		if !doc.VisibleLegacyOrganization(actor) {
			return nil, organizationDomainEntity.ErrNotFound
		}
		return &organizationID, nil
	})
	if errors.Is(err, organizationDomainEntity.ErrNotFound) {
		return "USER", actor, nil
	}
	if err != nil {
		return "", "", err
	}
	if selected == nil || *selected == "" {
		return "USER", actor, nil
	}
	return "ORGANIZATION", *selected, nil
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

func (s Store) RunOwnershipTransfer(ctx context.Context, organizationID string, transfer func(organizationapp.OwnershipWork) (*organizationDomainEntity.Organization, error)) (*organizationDomainEntity.Organization, error) {
	return orgstore.Run(ctx, s.DB, s.Channel, func(tx *firestore.Transaction, root *firestore.DocumentRef) (*organizationDomainEntity.Organization, error) {
		if uuid.Validate(organizationID) != nil || len(organizationID) != 36 {
			return nil, organizationDomainEntity.ErrInvalid
		}
		return transfer(ownershipWork{tx: tx, ref: root.Collection("organizations").Doc(organizationID)})
	})
}

type ownershipWork struct {
	tx  *firestore.Transaction
	ref *firestore.DocumentRef
}

func (w ownershipWork) Load() (organizationDomainEntity.Organization, []string, bool, error) {
	doc, err := orgstore.Read(w.tx, w.ref)
	if err != nil {
		return organizationDomainEntity.Organization{}, nil, false, err
	}
	org := organizationDomainEntity.Organization{ID: w.ref.ID, Name: doc.Name, OwnerUserID: doc.Owner}
	return org, doc.Members, doc.LegacyOrganizationEnabled(), nil
}

func (w ownershipWork) SaveOwner(previous, successor string) error {
	return w.tx.Update(w.ref, []firestore.Update{
		{Path: "ownerUserId", Value: successor},
		{Path: "memberUserIds", Value: firestore.ArrayUnion(previous)},
		{Path: "membershipOwnerVersion", Value: firestore.Increment(1)},
	})
}
