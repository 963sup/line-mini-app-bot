package organizationdb

import (
	"context"
	"errors"
	"regexp"
	"slices"
	"strings"

	"cloud.google.com/go/firestore"
	groupdomain "github.com/963sup/line-mini-app-bot/internal/business/group"
	"github.com/963sup/line-mini-app-bot/internal/business/organization"
	"github.com/963sup/line-mini-app-bot/internal/httpapi/generated"
	"github.com/963sup/line-mini-app-bot/internal/persistence/firestorekey"
	"github.com/google/uuid"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type Store struct {
	DB      *firestore.Client
	Channel string
	Groups  *groupdomain.Service
}

func run[T any](s Store, ctx context.Context, op func(*firestore.Transaction, *firestore.DocumentRef) (*T, error)) (*T, error) {
	var result *T
	err := s.DB.RunTransaction(ctx, func(_ context.Context, tx *firestore.Transaction) error {
		var e error
		result, e = op(tx, s.DB.Collection("channels").Doc(s.Channel))
		return e
	})
	if status.Code(err) == codes.NotFound {
		return nil, organization.ErrNotFound
	}
	return result, err
}

type entity struct {
	Name       string   `firestore:"name"`
	Owner      string   `firestore:"ownerUserId"`
	Members    []string `firestore:"memberUserIds"`
	Enterprise string   `firestore:"enterpriseId"`
	Managers   []string `firestore:"managerUserIds"`
	Scope      string   `firestore:"managementScope"`
	Manager    string   `firestore:"managementId"`
}

var userID = regexp.MustCompile("^U[0-9a-f]{32}$")

func read(tx *firestore.Transaction, ref *firestore.DocumentRef) (entity, error) {
	snap, e := tx.Get(ref)
	if e != nil {
		return entity{}, e
	}
	var d entity
	e = snap.DataTo(&d)
	return d, e
}

func (s Store) Context(ctx context.Context, user string, preference string, write bool) (*organization.Context, error) {
	return run(s, ctx, func(tx *firestore.Transaction, root *firestore.DocumentRef) (*organization.Context, error) {
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
			var o entity
			if err := d.DataTo(&o); err != nil {
				return nil, err
			}
			if _, ok := labels[o.Enterprise]; !ok && o.Enterprise != "" {
				ent, err := read(tx, root.Collection("enterprises").Doc(o.Enterprise))
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
		if write {
			if result.Selected == nil || result.Selected.ID != selected {
				return nil, organization.ErrNotFound
			}
			return &result, tx.Set(pref, map[string]any{"organizationId": selected})
		}
		return &result, nil
	})
}

func (s Store) Create(ctx context.Context, user string, name string) (*generated.Organization, error) {
	name = strings.TrimSpace(name)
	id := uuid.NewString()
	return run(s, ctx, func(tx *firestore.Transaction, root *firestore.DocumentRef) (*generated.Organization, error) {
		if name == "" || len([]rune(name)) > 120 {
			return nil, organization.ErrInvalid
		}
		return &generated.Organization{OrganizationId: id, OwnerUserId: user, Name: name}, tx.Create(root.Collection("organizations").Doc(id), map[string]any{"ownerUserId": user, "createdBy": user, "name": name})
	})
}

func (s Store) Transfer(ctx context.Context, user string, organizationId, target string) (*generated.Organization, error) {
	return run(s, ctx, func(tx *firestore.Transaction, root *firestore.DocumentRef) (*generated.Organization, error) {
		if uuid.Validate(organizationId) != nil || !userID.MatchString(target) {
			return nil, organization.ErrInvalid
		}
		ref := root.Collection("organizations").Doc(organizationId)
		org, e := read(tx, ref)
		if e != nil {
			return nil, e
		}
		if org.Owner != user {
			return nil, organization.ErrNotFound
		}
		if target != user && !slices.Contains(org.Members, target) {
			return nil, organization.ErrInvalid
		}
		return &generated.Organization{OrganizationId: organizationId, Name: org.Name, OwnerUserId: target}, tx.Update(ref, []firestore.Update{{Path: "ownerUserId", Value: target}})
	})
}

func (s Store) Manage(ctx context.Context, user string, teamId, scope, managementId string) error {
	_, err := run(s, ctx, func(tx *firestore.Transaction, root *firestore.DocumentRef) (*struct{}, error) {
		if !groupdomain.ValidID(teamId) || uuid.Validate(managementId) != nil || (scope != "organization" && scope != "enterprise") {
			return nil, organization.ErrInvalid
		}
		target, e := read(tx, root.Collection(scope+"s").Doc(managementId))
		if e != nil {
			return nil, e
		}
		if target.Owner != user && !slices.Contains(target.Managers, user) {
			return nil, organization.ErrNotFound
		}
		ref := root.Collection("teams").Doc(teamId)
		team, e := read(tx, ref)
		if e != nil {
			return nil, e
		}
		if !slices.Contains(team.Managers, user) {
			return nil, organization.ErrNotFound
		}
		if team.Scope != "" && (team.Scope != scope || team.Manager != managementId) {
			return nil, organization.ErrConflict
		}
		return nil, tx.Update(ref, []firestore.Update{{Path: "managementScope", Value: scope}, {Path: "managementId", Value: managementId}})
	})
	return err
}

func (s Store) Assign(ctx context.Context, user string, id, teamId string, add bool) error {
	_, err := run(s, ctx, func(tx *firestore.Transaction, root *firestore.DocumentRef) (*struct{}, error) {
		if uuid.Validate(id) != nil || !groupdomain.ValidID(teamId) {
			return nil, organization.ErrInvalid
		}
		ref := root.Collection("organizations").Doc(id)
		o, e := read(tx, ref)
		if e != nil {
			return nil, e
		}
		team, e := read(tx, root.Collection("teams").Doc(teamId))
		if e != nil {
			return nil, e
		}
		if o.Enterprise == "" || team.Scope != "enterprise" || team.Manager != o.Enterprise {
			return nil, organization.ErrNotFound
		}
		ent, e := read(tx, root.Collection("enterprises").Doc(o.Enterprise))
		if e != nil {
			return nil, e
		}
		if ent.Owner != user && !slices.Contains(ent.Managers, user) {
			return nil, organization.ErrNotFound
		}
		link := ref.Collection("enterpriseTeams").Doc(teamId)
		if add {
			return nil, tx.Set(link, map[string]any{"teamId": teamId})
		}
		return nil, tx.Delete(link)
	})
	return err
}

func (s Store) List(ctx context.Context, user string, organizationId string) (*generated.OrganizationTeamsResponse, error) {
	return run(s, ctx, func(tx *firestore.Transaction, root *firestore.DocumentRef) (*generated.OrganizationTeamsResponse, error) {
		if uuid.Validate(organizationId) != nil {
			return nil, organization.ErrInvalid
		}
		ref := root.Collection("organizations").Doc(organizationId)
		org, e := read(tx, ref)
		if e != nil {
			return nil, e
		}
		if org.Owner != user && !slices.Contains(org.Members, user) {
			return nil, organization.ErrNotFound
		}
		teams := root.Collection("teams")
		local, e := tx.Documents(teams.Where("managementId", "==", organizationId)).GetAll()
		if e != nil {
			return nil, e
		}
		ids := map[string]string{}
		for _, doc := range local {
			var team entity
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
			team, e := read(tx, teams.Doc(doc.Ref.ID))
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
		result := generated.OrganizationTeamsResponse{Teams: []generated.OrganizationTeamOption{}}
		for id, scope := range ids {
			if s.Groups == nil {
				return nil, groupdomain.ErrUnavailable
			}
			g, e := s.Groups.Get(ctx, user, id)
			if errors.Is(e, groupdomain.ErrNotFound) {
				continue
			}
			if e != nil {
				return nil, e
			}
			result.Teams = append(result.Teams, generated.OrganizationTeamOption{TeamId: id, Name: g.Name, ManagementScope: scope})
		}
		slices.SortFunc(result.Teams, func(a, b generated.OrganizationTeamOption) int { return strings.Compare(a.TeamId, b.TeamId) })
		return &result, nil
	})
}
