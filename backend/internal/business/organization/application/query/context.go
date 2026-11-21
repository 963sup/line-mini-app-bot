package query

import (
	"slices"
	"strings"

	organizationApplicationDto "github.com/963sup/line-mini-app-bot/internal/business/organization/application/dto"
)

// BuildContext orders authorized organizations and restores a still-visible preference.
func BuildContext(visible []organizationApplicationDto.VisibleOrganization, preference string) organizationApplicationDto.Context {
	rows := slices.Clone(visible)
	slices.SortFunc(rows, func(a, b organizationApplicationDto.VisibleOrganization) int {
		if (a.EnterpriseID == "") != (b.EnterpriseID == "") {
			if a.EnterpriseID == "" {
				return 1
			}
			return -1
		}
		return slices.Compare([]string{a.EnterpriseName, a.EnterpriseID, a.Name, a.ID}, []string{b.EnterpriseName, b.EnterpriseID, b.Name, b.ID})
	})
	result := organizationApplicationDto.Context{Groups: []organizationApplicationDto.Group{}, CanSwitch: len(rows) > 1}
	for i, r := range rows {
		if i == 0 || rows[i-1].EnterpriseID != r.EnterpriseID {
			g := organizationApplicationDto.Group{Label: "獨立組織"}
			if r.EnterpriseID != "" {
				g.EnterpriseID = &r.EnterpriseID
				g.Label = r.EnterpriseName
			}
			result.Groups = append(result.Groups, g)
		}
		g := &result.Groups[len(result.Groups)-1]
		o := organizationApplicationDto.Option{ID: r.ID, Name: r.Name}
		if i > 0 && rows[i-1].EnterpriseID == r.EnterpriseID && rows[i-1].Name == r.Name || i+1 < len(rows) && rows[i+1].EnterpriseID == r.EnterpriseID && rows[i+1].Name == r.Name {
			o.Detail = &r.ID
		}
		g.Organizations = append(g.Organizations, o)
		if i == 0 || r.ID == preference {
			result.Selected = &organizationApplicationDto.Selected{ID: r.ID, Name: r.Name, GroupLabel: g.Label}
		}
	}
	return result
}

// BuildContextWithNavigation combines the linked and directly visible enterprises.
func BuildContextWithNavigation(visible []organizationApplicationDto.VisibleOrganization, preference string, enterprises map[string]string, canCreate bool) organizationApplicationDto.Context {
	result := BuildContext(visible, preference)
	labels := make(map[string]string, len(enterprises))
	for _, row := range visible {
		if row.EnterpriseID != "" {
			labels[row.EnterpriseID] = row.EnterpriseName
		}
	}
	for id, name := range enterprises {
		labels[id] = name
	}
	result.CanCreateEnterprise = canCreate
	result.Enterprises = make([]organizationApplicationDto.EnterpriseLink, 0, len(labels))
	for id, name := range labels {
		result.Enterprises = append(result.Enterprises, organizationApplicationDto.EnterpriseLink{ID: id, Name: name})
	}
	slices.SortFunc(result.Enterprises, func(a, b organizationApplicationDto.EnterpriseLink) int {
		if a.Name == b.Name {
			return strings.Compare(a.ID, b.ID)
		}
		return strings.Compare(a.Name, b.Name)
	})
	return result
}
