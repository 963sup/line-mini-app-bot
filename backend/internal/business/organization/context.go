// Package organization models the workspaces available to a signed-in user.
package organization

import "slices"

// VisibleOrganization is supplied after membership and owner authorization.
type VisibleOrganization struct {
	ID, Name, EnterpriseID, EnterpriseName string
}

type Option struct {
	ID     string  `json:"organizationId"`
	Name   string  `json:"name"`
	Detail *string `json:"detail"`
}

type Group struct {
	EnterpriseID  *string  `json:"enterpriseId"`
	Label         string   `json:"label"`
	Organizations []Option `json:"organizations"`
}

type Selected struct {
	ID         string `json:"organizationId"`
	Name       string `json:"name"`
	GroupLabel string `json:"groupLabel"`
}

type Context struct {
	Groups    []Group   `json:"groups"`
	Selected  *Selected `json:"selected"`
	CanSwitch bool      `json:"canSwitch"`
}

// BuildContext orders authorized organizations and restores a still-visible preference.
func BuildContext(visible []VisibleOrganization, preference string) Context {
	rows := slices.Clone(visible)
	slices.SortFunc(rows, func(a, b VisibleOrganization) int {
		if (a.EnterpriseID == "") != (b.EnterpriseID == "") {
			if a.EnterpriseID == "" {
				return 1
			}
			return -1
		}
		return slices.Compare([]string{a.EnterpriseName, a.EnterpriseID, a.Name, a.ID}, []string{b.EnterpriseName, b.EnterpriseID, b.Name, b.ID})
	})
	result := Context{Groups: []Group{}, CanSwitch: len(rows) > 1}
	for i, r := range rows {
		if i == 0 || rows[i-1].EnterpriseID != r.EnterpriseID {
			g := Group{Label: "獨立組織"}
			if r.EnterpriseID != "" {
				g.EnterpriseID = &r.EnterpriseID
				g.Label = r.EnterpriseName
			}
			result.Groups = append(result.Groups, g)
		}
		g := &result.Groups[len(result.Groups)-1]
		o := Option{ID: r.ID, Name: r.Name}
		if i > 0 && rows[i-1].EnterpriseID == r.EnterpriseID && rows[i-1].Name == r.Name || i+1 < len(rows) && rows[i+1].EnterpriseID == r.EnterpriseID && rows[i+1].Name == r.Name {
			o.Detail = &r.ID
		}
		g.Organizations = append(g.Organizations, o)
		if i == 0 || r.ID == preference {
			result.Selected = &Selected{r.ID, r.Name, g.Label}
		}
	}
	return result
}
