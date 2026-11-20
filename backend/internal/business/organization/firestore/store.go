package organizationdb

import (
	"cloud.google.com/go/firestore"
	"context"
	enterprisedb "github.com/963sup/line-mini-app-bot/internal/business/enterprise/firestore"
	"github.com/963sup/line-mini-app-bot/internal/business/internal/orgstore"
	"github.com/963sup/line-mini-app-bot/internal/business/organization"

	"github.com/963sup/line-mini-app-bot/internal/persistence/firestorekey"
	"github.com/google/uuid"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"slices"
	"strings"
)

type Store struct {
	DB      *firestore.Client
	Channel string
}

func (s Store) Context(ctx context.Context, user string, preference string, write bool) (*organization.Context, error) {
	return orgstore.Run(ctx, s.DB, s.Channel, func(tx *firestore.Transaction, root *firestore.DocumentRef) (*organization.Context, error) {
		if write && uuid.Validate(preference) != nil {
			return nil, organization.ErrInvalid
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
		rows := make([]organization.VisibleOrganization, 0, len(docs))
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
			rows = append(rows, organization.VisibleOrganization{ID: d.Ref.ID, Name: o.Name, EnterpriseID: o.Enterprise, EnterpriseName: labels[o.Enterprise]})
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
		result := organization.BuildContext(rows, selected)
		if err := enterprisedb.PopulateNavigation(tx, root, user, labels, &result); err != nil {
			return nil, err
		}
		if write {
			if result.Selected == nil || result.Selected.ID != selected {
				return nil, organization.ErrNotFound
			}
			return &result, tx.Set(pref, map[string]any{"organizationId": selected})
		}
		return &result, nil
	})
}

func (s Store) Create(ctx context.Context, user string, name string) (*organization.Organization, error) {
	id := uuid.NewString()
	return orgstore.Run(ctx, s.DB, s.Channel, func(tx *firestore.Transaction, root *firestore.DocumentRef) (*organization.Organization, error) {
		return createOrganization(tx, root, user, id, strings.TrimSpace(name))
	})
}

func createOrganization(tx *firestore.Transaction, root *firestore.DocumentRef, user, id, name string) (*organization.Organization, error) {
	if name == "" || len([]rune(name)) > 120 {
		return nil, organization.ErrInvalid
	}
	return &organization.Organization{ID: id, OwnerUserID: user, Name: name}, tx.Create(root.Collection("organizations").Doc(id), map[string]any{"ownerUserId": user, "createdBy": user, "name": name})
}

func (s Store) Transfer(ctx context.Context, user string, organizationId, target string) (*organization.Organization, error) {
	return orgstore.Run(ctx, s.DB, s.Channel, func(tx *firestore.Transaction, root *firestore.DocumentRef) (*organization.Organization, error) {
		if uuid.Validate(organizationId) != nil || !orgstore.UserID.MatchString(target) {
			return nil, organization.ErrInvalid
		}
		ref := root.Collection("organizations").Doc(organizationId)
		org, e := orgstore.Read(tx, ref)
		if e != nil {
			return nil, e
		}
		if org.Owner != user {
			return nil, organization.ErrNotFound
		}
		if target != user && !slices.Contains(org.Members, target) {
			return nil, organization.ErrInvalid
		}
		return &organization.Organization{ID: organizationId, Name: org.Name, OwnerUserID: target}, tx.Update(ref, []firestore.Update{{Path: "ownerUserId", Value: target}})
	})
}
