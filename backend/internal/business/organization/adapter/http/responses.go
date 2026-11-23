package organizationhttp

import (
	organizationdto "github.com/963sup/line-mini-app-bot/internal/business/organization/application/dto"
	organizationDomainEntity "github.com/963sup/line-mini-app-bot/internal/business/organization/domain/entity"
)

import (
	"github.com/963sup/line-mini-app-bot/internal/transport/http/api/generated"
)

func organizationResponse(value *organizationDomainEntity.Organization, err error) (any, error) {
	if err != nil || value == nil {
		return nil, err
	}
	return generated.Organization{Name: value.Name, OrganizationId: value.ID, OwnerUserId: value.OwnerUserID}, nil
}

// contextResponse keeps the public HTTP schema outside the application DTOs.
func contextResponse(value *organizationdto.Context, err error) (any, error) {
	if err != nil || value == nil {
		return nil, err
	}
	result := generated.OrganizationContext{
		CanCreateEnterprise: value.CanCreateEnterprise, CanSwitch: value.CanSwitch,
	}
	if value.Enterprises != nil {
		result.Enterprises = make([]generated.EnterpriseLink, 0, len(value.Enterprises))
	}
	if value.Groups != nil {
		result.Groups = make([]generated.OrganizationGroup, 0, len(value.Groups))
	}
	for _, e := range value.Enterprises {
		result.Enterprises = append(result.Enterprises, generated.EnterpriseLink{Id: e.ID, Name: e.Name})
	}
	for _, g := range value.Groups {
		group := generated.OrganizationGroup{EnterpriseId: g.EnterpriseID, Label: g.Label}
		if g.Organizations != nil {
			group.Organizations = make([]generated.OrganizationOption, 0, len(g.Organizations))
		}
		for _, o := range g.Organizations {
			group.Organizations = append(group.Organizations, generated.OrganizationOption{OrganizationId: o.ID, Name: o.Name, Detail: o.Detail})
		}
		result.Groups = append(result.Groups, group)
	}
	if value.Selected != nil {
		result.Selected = &generated.SelectedOrganization{OrganizationId: value.Selected.ID, Name: value.Selected.Name, GroupLabel: value.Selected.GroupLabel}
	}
	return result, nil
}
