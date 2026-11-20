package enterprisedb

import (
	organizationApplicationDto "github.com/963sup/line-mini-app-bot/internal/business/organization/application/dto"
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

	groupport "github.com/963sup/line-mini-app-bot/internal/business/group/application/port"
	groupquery "github.com/963sup/line-mini-app-bot/internal/business/group/application/query"
	groupentity "github.com/963sup/line-mini-app-bot/internal/business/group/domain/entity"
	"github.com/963sup/line-mini-app-bot/internal/business/internal/adapter/firestore"

	"github.com/963sup/line-mini-app-bot/internal/persistence/firestorekey"
	"github.com/google/uuid"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type Store struct {
	DB      *firestore.Client
	Channel string
	Groups  *groupquery.Service
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

func PopulateNavigation(tx *firestore.Transaction, root *firestore.DocumentRef, user string, labels map[string]string, c *organizationApplicationDto.Context) error {
	g, _, e := Provisioning(tx, root, user)
	if e != nil {
		return e
	}
	c.CanCreateEnterprise = policy.CanProvision(g.Enabled, g.EnterpriseID, false)
	// Use individual indexed queries: Firestore does not support multiple array-contains filters in one disjunction in all deployments.
	for _, q := range []firestore.Query{root.Collection("enterprises").Where("ownerUserId", "==", user), root.Collection("enterprises").Where("ownerUserIds", "array-contains", user), root.Collection("enterprises").Where("managerUserIds", "array-contains", user)} {
		docs, e := tx.Documents(q).GetAll()
		if e != nil {
			return e
		}
		for _, doc := range docs {
			var ent orgstore.Document
			if e := doc.DataTo(&ent); e != nil {
				return e
			}
			labels[doc.Ref.ID] = ent.Name
		}
	}
	c.Enterprises = []organizationApplicationDto.EnterpriseLink{}
	for id, name := range labels {
		c.Enterprises = append(c.Enterprises, organizationApplicationDto.EnterpriseLink{ID: id, Name: name})
	}
	slices.SortFunc(c.Enterprises, func(a, b organizationApplicationDto.EnterpriseLink) int {
		if a.Name == b.Name {
			return strings.Compare(a.ID, b.ID)
		}
		return strings.Compare(a.Name, b.Name)
	})
	return nil
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
		if !existing.IsOwner(user) {
			return nil, organizationDomainEntity.ErrForbidden
		}
		return &dto.Link{ID: g.EnterpriseID, Name: existing.Name}, nil
	}
	if e := tx.Create(root.Collection("enterprises").Doc(id), map[string]any{"name": name, "ownerUserIds": []string{user}, "createdBy": user}); e != nil {
		return nil, e
	}
	return &dto.Link{ID: id, Name: name}, tx.Update(ref, []firestore.Update{{Path: "enterpriseId", Value: id}})
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
		owner := ent.IsOwner(user)
		admin := owner || slices.Contains(ent.Managers, user)
		view := dto.Detail{ID: id, Name: ent.Name, CanManage: owner, Organizations: []dto.Link{}, Teams: []dto.Link{}}
		docs, e := tx.Documents(root.Collection("organizations").Where("enterpriseId", "==", id)).GetAll()
		if e != nil {
			return nil, e
		}
		for _, d := range docs {
			var org orgstore.Document
			if e := d.DataTo(&org); e != nil {
				return nil, e
			}
			if admin || org.Owner == user || slices.Contains(org.Members, user) {
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
			if team.Scope != "enterprise" {
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

func (s Store) UpdateEnterprise(ctx context.Context, user, id, target string, addOwner bool) (*dto.Link, error) {
	if uuid.Validate(id) != nil || (addOwner && !orgstore.UserID.MatchString(target)) || (!addOwner && uuid.Validate(target) != nil) {
		return nil, organizationDomainEntity.ErrInvalid
	}
	return orgstore.Run(ctx, s.DB, s.Channel, func(tx *firestore.Transaction, root *firestore.DocumentRef) (*dto.Link, error) {
		ref := root.Collection("enterprises").Doc(id)
		ent, e := orgstore.Read(tx, ref)
		if e != nil {
			return nil, e
		}
		if !ent.IsOwner(user) {
			return nil, organizationDomainEntity.ErrNotFound
		}
		link := &dto.Link{ID: id, Name: ent.Name}
		if addOwner {
			return link, tx.Update(ref, []firestore.Update{{Path: "ownerUserIds", Value: firestore.ArrayUnion(target)}})
		}
		orgRef := root.Collection("organizations").Doc(target)
		org, e := orgstore.Read(tx, orgRef)
		if e != nil {
			return nil, e
		}
		if org.Owner != user {
			return nil, organizationDomainEntity.ErrNotFound
		}
		if org.Enterprise != "" && org.Enterprise != id {
			return nil, organizationDomainEntity.ErrConflict
		}
		return link, tx.Update(orgRef, []firestore.Update{{Path: "enterpriseId", Value: id}})
	})
}
