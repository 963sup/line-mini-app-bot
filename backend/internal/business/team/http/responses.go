package teamhttp

import (
	"github.com/963sup/line-mini-app-bot/internal/business/team"
	"github.com/963sup/line-mini-app-bot/internal/httpapi/generated"
)

func teamResponse(value *team.Listing, err error) (any, error) {
	if err != nil || value == nil {
		return nil, err
	}
	teams := make([]generated.OrganizationTeamOption, 0, len(value.Teams))
	for _, option := range value.Teams {
		teams = append(teams, generated.OrganizationTeamOption{TeamId: option.ID, Name: option.Name, ManagementScope: option.ManagementScope})
	}
	return generated.OrganizationTeamsResponse{Teams: teams}, nil
}
