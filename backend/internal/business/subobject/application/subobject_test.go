package application

import (
	"errors"
	"testing"

	objectDomainEntity "github.com/963sup/line-mini-app-bot/internal/business/object/domain/entity"
	subobjectDomainEntity "github.com/963sup/line-mini-app-bot/internal/business/subobject/domain/entity"
)

func pointer(v string) *string { return &v }
func TestIndependentExactProgressAndConversions(t *testing.T) {
	c, e := subobjectDomainEntity.Normalize(subobjectDomainEntity.Content{Name: "Lay", Type: "work", Quantity: "999999999999999999.999999", Unit: "M", CompletedQuantity: "999999999999999999.999998"})
	if e != nil {
		t.Fatal(e)
	}
	status, e := subobjectDomainEntity.Status(c)
	if e != nil || status != objectDomainEntity.StatusInProgress {
		t.Fatal(status, e)
	}
	j := subobjectDomainEntity.Subobject{Content: c, CreatorID: "owner", Status: status, Version: 1}
	u := subobjectDomainEntity.Update{Content: c, ExpectedVersion: 1}
	same, e := subobjectDomainEntity.ApplyUpdate(j, "owner", u)
	if e != nil || same.Version != 1 || same.CompletedQuantity != c.CompletedQuantity {
		t.Fatal("noop reset", same, e)
	}
	u.Type = "non_work"
	if _, e = subobjectDomainEntity.ApplyUpdate(j, "owner", u); !errors.Is(e, subobjectDomainEntity.ErrProgressRemoval) {
		t.Fatal(e)
	}
	u.Type = "work"
	u.Quantity = "1"
	if _, e = subobjectDomainEntity.ApplyUpdate(j, "owner", u); !errors.Is(e, subobjectDomainEntity.ErrInput) {
		t.Fatal("shrink", e)
	}
	u.CompletedQuantity = pointer("1")
	next, e := subobjectDomainEntity.ApplyUpdate(j, "owner", u)
	if e != nil || next.Status != objectDomainEntity.StatusCompleted || next.Version != 2 {
		t.Fatal(next, e)
	}
	if _, e = subobjectDomainEntity.ApplyUpdate(j, "other", u); !errors.Is(e, subobjectDomainEntity.ErrForbidden) {
		t.Fatal(e)
	}
	u.ExpectedVersion = 2
	if _, e = subobjectDomainEntity.ApplyUpdate(j, "owner", u); !errors.Is(e, subobjectDomainEntity.ErrConflict) {
		t.Fatal(e)
	}
	j.CompletedQuantity = "0"
	j.Status = objectDomainEntity.StatusNotStarted
	u = subobjectDomainEntity.Update{Content: j.Content, ExpectedVersion: 1}
	u.Type = "non_work"
	non, e := subobjectDomainEntity.ApplyUpdate(j, "owner", u)
	if e != nil || non.CompletedQuantity != "" {
		t.Fatal(non, e)
	}
	u = subobjectDomainEntity.Update{Content: non.Content, ExpectedVersion: 2}
	u.Type = "work"
	if _, e = subobjectDomainEntity.ApplyUpdate(non, "owner", u); !errors.Is(e, subobjectDomainEntity.ErrInput) {
		t.Fatal("implicit conversion", e)
	}
	for _, bad := range []subobjectDomainEntity.Content{{Name: "x", Type: "legacy", Quantity: "1", Unit: "M"}, {Name: "x", Type: "work", Quantity: "0", Unit: "M"}, {Name: "x", Type: "work", Quantity: "1", Unit: "M", CompletedQuantity: "NaN"}, {Name: "x", Type: "non_work", Quantity: "1", Unit: "M", CompletedQuantity: "0"}} {
		if _, e = subobjectDomainEntity.Normalize(bad); !errors.Is(e, subobjectDomainEntity.ErrInput) {
			t.Fatal(bad, e)
		}
	}
}
