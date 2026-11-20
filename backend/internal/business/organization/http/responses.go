package organizationhttp

import (
	"github.com/963sup/line-mini-app-bot/internal/business/enterprise"
	"github.com/963sup/line-mini-app-bot/internal/business/organization"
	"github.com/963sup/line-mini-app-bot/internal/business/team"
	"github.com/963sup/line-mini-app-bot/internal/httpapi/generated"
)

func organizationResponse(value *organization.Organization, err error) (any, error) {
	if err != nil || value == nil {
		return nil, err
	}
	return generated.Organization{Name: value.Name, OrganizationId: value.ID, OwnerUserId: value.OwnerUserID}, nil
}

func enterpriseLink(value enterprise.Link) generated.EnterpriseLink {
	return generated.EnterpriseLink{Id: value.ID, Name: value.Name}
}

func enterpriseLinkResponse(value *enterprise.Link, err error) (any, error) {
	if err != nil || value == nil {
		return nil, err
	}
	return enterpriseLink(*value), nil
}

func enterpriseLinks(values []enterprise.Link) []generated.EnterpriseLink {
	result := make([]generated.EnterpriseLink, 0, len(values))
	for _, value := range values {
		result = append(result, enterpriseLink(value))
	}
	return result
}

func enterpriseResponse(value *enterprise.Detail, err error) (any, error) {
	if err != nil || value == nil {
		return nil, err
	}
	return generated.Enterprise{
		Id: value.ID, Name: value.Name, CanManage: value.CanManage,
		Organizations: enterpriseLinks(value.Organizations), Teams: enterpriseLinks(value.Teams),
	}, nil
}

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
