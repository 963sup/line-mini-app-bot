package teamhttp

import (
	teamApplicationDto "github.com/963sup/line-mini-app-bot/internal/business/team/application/dto"
)

import (
	"github.com/963sup/line-mini-app-bot/internal/transport/http/api/generated"
)

func teamResponse(value *teamApplicationDto.Listing, err error) (any, error) {
	if err != nil || value == nil {
		return nil, err
	}
	teams := make([]generated.OrganizationTeamOption, 0, len(value.Teams))
	for _, option := range value.Teams {
		teams = append(teams, generated.OrganizationTeamOption{AuditHistoryStatus: generated.AuditHistoryStatusUnavailable, TeamId: option.ID, Name: option.Name, ManagementScope: option.ManagementScope})
	}
	return generated.OrganizationTeamsResponse{Teams: teams, CanBindTeam: value.CanBindTeam}, nil
}
