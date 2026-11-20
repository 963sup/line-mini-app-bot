package organizationhttp

import (
	"github.com/963sup/line-mini-app-bot/internal/business/organization"
	"github.com/963sup/line-mini-app-bot/internal/httpapi/generated"
)

func organizationResponse(value *organization.Organization, err error) (any, error) {
	if err != nil || value == nil {
		return nil, err
	}
	return generated.Organization{Name: value.Name, OrganizationId: value.ID, OwnerUserId: value.OwnerUserID}, nil
}
