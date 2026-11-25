package teamdb

import (
	"cloud.google.com/go/firestore"
	dto "github.com/963sup/line-mini-app-bot/internal/business/enterprise/application/dto"
	orgstore "github.com/963sup/line-mini-app-bot/internal/business/internal/adapter/firestore"
)

func (s Store) EnterpriseTeams(tx *firestore.Transaction, root *firestore.DocumentRef, user, id string, admin bool, organizations []dto.Link) ([]dto.TeamCandidate, error) {
	var candidates []dto.TeamCandidate
	assigned := map[string]bool{}
	if !admin {
		for _, org := range organizations {
			links, err := tx.Documents(root.Collection("organizations").Doc(org.ID).Collection("enterpriseTeams").Query).GetAll()
			if err != nil {
				return nil, err
			}
			for _, link := range links {
				assigned[link.Ref.ID] = true
			}
		}
	}
	docs, e := tx.Documents(root.Collection("teams").Where("managementId", "==", id)).GetAll()
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
				candidates = append(candidates, dto.TeamCandidate{Link: dto.Link{ID: d.Ref.ID, Name: team.Name}})
			}
			continue
		case "", "legacy":
		default:
			continue
		}
		if !admin && !assigned[d.Ref.ID] {
			continue
		}
		candidates = append(candidates, dto.TeamCandidate{CheckLINE: !admin, Link: dto.Link{ID: d.Ref.ID, Name: team.Name}})
	}

	return candidates, nil
}
