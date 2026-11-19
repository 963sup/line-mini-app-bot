package organization

import "testing"

func TestContext(t *testing.T) {
	rows := []VisibleOrganization{{ID: "3", Name: "A"}, {ID: "2", Name: "A", EnterpriseID: "E", EnterpriseName: "HQ"}, {ID: "1", Name: "A", EnterpriseID: "E", EnterpriseName: "HQ"}}
	c := BuildContext(rows, "3")
	if !c.CanSwitch || len(c.Groups) != 2 || c.Selected.ID != "3" || c.Selected.GroupLabel != "獨立組織" {
		t.Fatal(c)
	}
	for _, o := range c.Groups[0].Organizations {
		if o.Detail == nil || *o.Detail != o.ID {
			t.Fatal(o)
		}
	}
	if BuildContext(rows, "revoked").Selected.ID != "1" {
		t.Fatal("unexpected fallback")
	}
	if empty := BuildContext(nil, "3"); empty.Groups == nil || empty.Selected != nil || empty.CanSwitch {
		t.Fatal(empty)
	}
}
