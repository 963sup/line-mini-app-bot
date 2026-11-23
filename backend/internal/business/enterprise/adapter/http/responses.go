package enterprisehttp

import (
	"errors"

	"github.com/963sup/line-mini-app-bot/internal/business/enterprise/application/dto"
	enterprise "github.com/963sup/line-mini-app-bot/internal/business/enterprise/domain"
	legacy "github.com/963sup/line-mini-app-bot/internal/business/organization/domain/entity"
	"github.com/963sup/line-mini-app-bot/internal/transport/http/api/generated"
)

func enterpriseLink(value dto.Link) generated.EnterpriseLink {
	return generated.EnterpriseLink{Id: value.ID, Name: value.Name}
}

func enterpriseLinkResponse(value *dto.Link, err error) (any, error) {
	if err != nil || value == nil {
		return nil, enterpriseResponseError(err)
	}
	return enterpriseLink(*value), nil
}

func enterpriseLinks(values []dto.Link) []generated.EnterpriseLink {
	result := make([]generated.EnterpriseLink, 0, len(values))
	for _, value := range values {
		result = append(result, enterpriseLink(value))
	}
	return result
}

func enterpriseResponse(value *dto.Detail, err error) (any, error) {
	if err != nil || value == nil {
		return nil, enterpriseResponseError(err)
	}
	teams := enterpriseLinks(value.Teams)
	status := generated.AuditHistoryStatusUnavailable
	for i := range teams {
		teams[i].AuditHistoryStatus = &status
	}
	return generated.Enterprise{
		AuditHistoryStatus: status,
		Id:                 value.ID, Name: value.Name, CanManage: value.CanManage, CanManageTeams: value.CanManageTeams,
		Organizations: enterpriseLinks(value.Organizations), AssignableOrganizations: enterpriseLinks(value.AssignableOrganizations), Teams: teams,
	}, nil
}

// Translate Enterprise-owned errors only at the existing HTTP compatibility
// boundary. Storage errors already using the legacy transport categories pass
// through; the core no longer depends on another context's error values.
func enterpriseResponseError(err error) error {
	switch {
	case errors.Is(err, enterprise.ErrNotFound):
		return errors.Join(legacy.ErrNotFound, err)
	case errors.Is(err, enterprise.ErrConflict):
		return errors.Join(legacy.ErrConflict, err)
	default:
		return err
	}
}
