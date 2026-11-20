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
	enterprisedb "github.com/963sup/line-mini-app-bot/internal/business/enterprise/adapter/firestore"
	"github.com/963sup/line-mini-app-bot/internal/business/internal/adapter/firestore"

	"github.com/963sup/line-mini-app-bot/internal/platform/firestorekey"
	"github.com/google/uuid"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type Store struct {
	DB      *firestore.Client
	Channel string
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
		result := organizationApplicationQuery.BuildContext(rows, selected)
		if err := enterprisedb.PopulateNavigation(tx, root, user, labels, &result); err != nil {
			return nil, err
		}
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
	return &organizationDomainEntity.Organization{ID: id, OwnerUserID: user, Name: name}, tx.Create(root.Collection("organizations").Doc(id), map[string]any{"ownerUserId": user, "createdBy": user, "name": name})
}

func (s Store) Transfer(ctx context.Context, user string, organizationId, target string) (*organizationDomainEntity.Organization, error) {
	return orgstore.Run(ctx, s.DB, s.Channel, func(tx *firestore.Transaction, root *firestore.DocumentRef) (*organizationDomainEntity.Organization, error) {
		if uuid.Validate(organizationId) != nil || !orgstore.UserID.MatchString(target) {
			return nil, organizationDomainEntity.ErrInvalid
		}
		ref := root.Collection("organizations").Doc(organizationId)
		org, e := orgstore.Read(tx, ref)
		if e != nil {
			return nil, e
		}
		if org.Owner != user {
			return nil, organizationDomainEntity.ErrNotFound
		}
		if target != user && !slices.Contains(org.Members, target) {
			return nil, organizationDomainEntity.ErrInvalid
		}
		return &organizationDomainEntity.Organization{ID: organizationId, Name: org.Name, OwnerUserID: target}, tx.Update(ref, []firestore.Update{{Path: "ownerUserId", Value: target}})
	})
}
