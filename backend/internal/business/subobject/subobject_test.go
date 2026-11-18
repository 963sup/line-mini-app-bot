package subobject

import (
	"errors"
	"testing"

	"github.com/963sup/line-mini-app-bot/internal/business/object"
)

func pointer(v string) *string { return &v }
func TestIndependentExactProgressAndConversions(t *testing.T) {
	c, e := Normalize(Content{Name: "Lay", Type: "work", Quantity: "999999999999999999.999999", Unit: "M", CompletedQuantity: "999999999999999999.999998"})
	if e != nil {
		t.Fatal(e)
	}
	status, e := Status(c)
	if e != nil || status != object.StatusInProgress {
		t.Fatal(status, e)
	}
	j := Subobject{Content: c, CreatorID: "owner", Status: status, Version: 1}
	u := Update{Content: c, ExpectedVersion: 1}
	same, e := ApplyUpdate(j, "owner", u)
	if e != nil || same.Version != 1 || same.CompletedQuantity != c.CompletedQuantity {
		t.Fatal("noop reset", same, e)
	}
	u.Type = "non_work"
	if _, e = ApplyUpdate(j, "owner", u); !errors.Is(e, ErrProgressRemoval) {
		t.Fatal(e)
	}
	u.Type = "work"
	u.Quantity = "1"
	if _, e = ApplyUpdate(j, "owner", u); !errors.Is(e, ErrInput) {
		t.Fatal("shrink", e)
	}
	u.CompletedQuantity = pointer("1")
	next, e := ApplyUpdate(j, "owner", u)
	if e != nil || next.Status != object.StatusCompleted || next.Version != 2 {
		t.Fatal(next, e)
	}
	if _, e = ApplyUpdate(j, "other", u); !errors.Is(e, ErrForbidden) {
		t.Fatal(e)
	}
	u.ExpectedVersion = 2
	if _, e = ApplyUpdate(j, "owner", u); !errors.Is(e, ErrConflict) {
		t.Fatal(e)
	}
	j.CompletedQuantity = "0"
	j.Status = object.StatusNotStarted
	u = Update{Content: j.Content, ExpectedVersion: 1}
	u.Type = "non_work"
	non, e := ApplyUpdate(j, "owner", u)
	if e != nil || non.CompletedQuantity != "" {
		t.Fatal(non, e)
	}
	u = Update{Content: non.Content, ExpectedVersion: 2}
	u.Type = "work"
	if _, e = ApplyUpdate(non, "owner", u); !errors.Is(e, ErrInput) {
		t.Fatal("implicit conversion", e)
	}
	for _, bad := range []Content{{Name: "x", Type: "legacy", Quantity: "1", Unit: "M"}, {Name: "x", Type: "work", Quantity: "0", Unit: "M"}, {Name: "x", Type: "work", Quantity: "1", Unit: "M", CompletedQuantity: "NaN"}, {Name: "x", Type: "non_work", Quantity: "1", Unit: "M", CompletedQuantity: "0"}} {
		if _, e = Normalize(bad); !errors.Is(e, ErrInput) {
			t.Fatal(bad, e)
		}
	}
}
