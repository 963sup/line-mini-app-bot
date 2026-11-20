package application

import (
	"testing"

	organizationApplicationDto "github.com/963sup/line-mini-app-bot/internal/business/organization/application/dto"
	organizationApplicationQuery "github.com/963sup/line-mini-app-bot/internal/business/organization/application/query"
)

func TestContext(t *testing.T) {
	rows := []organizationApplicationDto.VisibleOrganization{{ID: "3", Name: "A"}, {ID: "2", Name: "A", EnterpriseID: "E", EnterpriseName: "HQ"}, {ID: "1", Name: "A", EnterpriseID: "E", EnterpriseName: "HQ"}}
	c := organizationApplicationQuery.BuildContext(rows, "3")
	if !c.CanSwitch || len(c.Groups) != 2 || c.Selected.ID != "3" || c.Selected.GroupLabel != "獨立組織" {
		t.Fatal(c)
	}
	for _, o := range c.Groups[0].Organizations {
		if o.Detail == nil || *o.Detail != o.ID {
			t.Fatal(o)
		}
	}
	if organizationApplicationQuery.BuildContext(rows, "revoked").Selected.ID != "1" {
		t.Fatal("unexpected fallback")
	}
	if empty := organizationApplicationQuery.BuildContext(nil, "3"); empty.Groups == nil || empty.Selected != nil || empty.CanSwitch {
		t.Fatal(empty)
	}
}
