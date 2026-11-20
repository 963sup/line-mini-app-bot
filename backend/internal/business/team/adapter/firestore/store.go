package teamdb

import (
	teamApplicationDto "github.com/963sup/line-mini-app-bot/internal/business/team/application/dto"
)

import (
	organizationDomainEntity "github.com/963sup/line-mini-app-bot/internal/business/organization/domain/entity"
)

import (
	"context"
	"errors"
	"slices"
	"strings"

	"cloud.google.com/go/firestore"

	groupport "github.com/963sup/line-mini-app-bot/internal/business/group/application/port"
	groupquery "github.com/963sup/line-mini-app-bot/internal/business/group/application/query"
	groupentity "github.com/963sup/line-mini-app-bot/internal/business/group/domain/entity"
	groupvalueobject "github.com/963sup/line-mini-app-bot/internal/business/group/domain/valueobject"
	"github.com/963sup/line-mini-app-bot/internal/business/internal/adapter/firestore"

	"github.com/google/uuid"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type Store struct {
	DB      *firestore.Client
	Channel string
	Groups  *groupquery.Service
}

func (s Store) Manage(ctx context.Context, user string, teamId, scope, managementId string) error {
	_, err := orgstore.Run(ctx, s.DB, s.Channel, func(tx *firestore.Transaction, root *firestore.DocumentRef) (*struct{}, error) {
		if !groupvalueobject.ValidID(teamId) || uuid.Validate(managementId) != nil || (scope != "organization" && scope != "enterprise") {
			return nil, organizationDomainEntity.ErrInvalid
		}
		target, e := orgstore.Read(tx, root.Collection(scope+"s").Doc(managementId))
		if e != nil {
			return nil, e
		}
		owner := target.Owner == user
		if scope == "enterprise" {
			owner = target.IsOwner(user)
		}
		if !owner && !slices.Contains(target.Managers, user) {
			return nil, organizationDomainEntity.ErrNotFound
		}
		ref := root.Collection("teams").Doc(teamId)
		team, e := orgstore.Read(tx, ref)
		if e != nil {
			return nil, e
		}
		if !slices.Contains(team.Managers, user) {
			return nil, organizationDomainEntity.ErrNotFound
		}
		if team.Scope != "" && (team.Scope != scope || team.Manager != managementId) {
			return nil, organizationDomainEntity.ErrConflict
		}
		return nil, tx.Update(ref, []firestore.Update{{Path: "managementScope", Value: scope}, {Path: "managementId", Value: managementId}})
	})
	return err
}

func (s Store) Assign(ctx context.Context, user string, id, teamId string, add bool) error {
	_, err := orgstore.Run(ctx, s.DB, s.Channel, func(tx *firestore.Transaction, root *firestore.DocumentRef) (*struct{}, error) {
		if uuid.Validate(id) != nil || !groupvalueobject.ValidID(teamId) {
			return nil, organizationDomainEntity.ErrInvalid
		}
		ref := root.Collection("organizations").Doc(id)
		o, e := orgstore.Read(tx, ref)
		if e != nil {
			return nil, e
		}
		team, e := orgstore.Read(tx, root.Collection("teams").Doc(teamId))
		if e != nil {
			return nil, e
		}
		if o.Enterprise == "" || team.Scope != "enterprise" || team.Manager != o.Enterprise {
			return nil, organizationDomainEntity.ErrNotFound
		}
		ent, e := orgstore.Read(tx, root.Collection("enterprises").Doc(o.Enterprise))
		if e != nil {
			return nil, e
		}
		if !ent.IsOwner(user) && !slices.Contains(ent.Managers, user) {
			return nil, organizationDomainEntity.ErrNotFound
		}
		link := ref.Collection("enterpriseTeams").Doc(teamId)
		if add {
			return nil, tx.Set(link, map[string]any{"teamId": teamId})
		}
		return nil, tx.Delete(link)
	})
	return err
}

func (s Store) List(ctx context.Context, user string, organizationId string) (*teamApplicationDto.Listing, error) {
	return orgstore.Run(ctx, s.DB, s.Channel, func(tx *firestore.Transaction, root *firestore.DocumentRef) (*teamApplicationDto.Listing, error) {
		if uuid.Validate(organizationId) != nil {
			return nil, organizationDomainEntity.ErrInvalid
		}
		ref := root.Collection("organizations").Doc(organizationId)
		org, e := orgstore.Read(tx, ref)
		if e != nil {
			return nil, e
		}
		if org.Owner != user && !slices.Contains(org.Members, user) {
			return nil, organizationDomainEntity.ErrNotFound
		}
		teams := root.Collection("teams")
		local, e := tx.Documents(teams.Where("managementId", "==", organizationId)).GetAll()
		if e != nil {
			return nil, e
		}
		ids := map[string]string{}
		for _, doc := range local {
			var team orgstore.Document
			if e := doc.DataTo(&team); e != nil {
				return nil, e
			}
			if team.Scope == "organization" {
				ids[doc.Ref.ID] = "organization"
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
			if org.Enterprise != "" && team.Scope == "enterprise" && team.Manager == org.Enterprise {
				ids[doc.Ref.ID] = "enterprise"
			}
		}
		result := teamApplicationDto.Listing{Teams: []teamApplicationDto.Option{}}
		for id, scope := range ids {
			if s.Groups == nil {
				return nil, groupport.ErrUnavailable
			}
			g, e := s.Groups.Get(ctx, user, id)
			if errors.Is(e, groupentity.ErrNotFound) {
				continue
			}
			if e != nil {
				return nil, e
			}
			result.Teams = append(result.Teams, teamApplicationDto.Option{ID: id, Name: g.Name, ManagementScope: scope})
		}
		slices.SortFunc(result.Teams, func(a, b teamApplicationDto.Option) int { return strings.Compare(a.ID, b.ID) })
		return &result, nil
	})
}
