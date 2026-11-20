package enterprisehttp

import (
	"github.com/963sup/line-mini-app-bot/internal/business/enterprise/application/dto"
	"github.com/963sup/line-mini-app-bot/internal/httpapi/generated"
)

func enterpriseLink(value dto.Link) generated.EnterpriseLink {
	return generated.EnterpriseLink{Id: value.ID, Name: value.Name}
}

func enterpriseLinkResponse(value *dto.Link, err error) (any, error) {
	if err != nil || value == nil {
		return nil, err
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
		return nil, err
	}
	return generated.Enterprise{
		Id: value.ID, Name: value.Name, CanManage: value.CanManage,
		Organizations: enterpriseLinks(value.Organizations), Teams: enterpriseLinks(value.Teams),
	}, nil
}
