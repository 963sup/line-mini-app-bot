package application

import (
	"errors"
	"testing"

	objectDomainEntity "github.com/963sup/line-mini-app-bot/internal/business/object/domain/entity"
)

func ptr(s string) *string { return &s }
func TestExactWorkStatus(t *testing.T) {
	for _, c := range []struct{ total, done, status string }{{"11", "0", objectDomainEntity.StatusNotStarted}, {"11", "5.5", objectDomainEntity.StatusInProgress}, {"11", "11.000000", objectDomainEntity.StatusCompleted}, {"999999999999999999.999999", "999999999999999999.999998", objectDomainEntity.StatusInProgress}, {"0.000001", "0.000001", objectDomainEntity.StatusCompleted}} {
		got, e := objectDomainEntity.WorkStatus(c.total, c.done)
		if e != nil || got != c.status {
			t.Fatal(c, got, e)
		}
	}
	for _, done := range []string{"-1", "12", "NaN", "Inf", "1e1", "1/2", "0.0000001", "1000000000000000000"} {
		if _, e := objectDomainEntity.WorkStatus("11", done); !errors.Is(e, objectDomainEntity.ErrInput) {
			t.Fatal(done, e)
		}
	}
}

func TestProgressConversionsAndAbsoluteUpdates(t *testing.T) {
	base := content()
	base.Quantity = "11"
	base.Type = "Engineering"
	current := objectDomainEntity.Object{Content: base, CreatorID: "A", Status: objectDomainEntity.StatusInProgress, Version: 3}
	update := objectDomainEntity.Update{Content: base, ExpectedVersion: 3}
	got, e := objectDomainEntity.ApplyUpdate(current, "A", update)
	if e != nil || got.Version != 3 || got.Status != objectDomainEntity.StatusInProgress {
		t.Fatal("legacy noop", got, e)
	}
	update.Status = objectDomainEntity.StatusCompleted
	got, e = objectDomainEntity.ApplyUpdate(current, "A", update)
	if e != nil || got.Status != objectDomainEntity.StatusCompleted || got.CompletedQuantity != "" {
		t.Fatal("legacy status", got, e)
	}
	update.Type = " work "
	update.Status = ""
	if _, e = objectDomainEntity.ApplyUpdate(current, "A", update); !errors.Is(e, objectDomainEntity.ErrInput) {
		t.Fatal("conversion silently defaulted", e)
	}
	update.CompletedQuantity = ptr("5.500000")
	got, e = objectDomainEntity.ApplyUpdate(current, "A", update)
	if e != nil || got.Status != objectDomainEntity.StatusInProgress || got.CompletedQuantity != "5.5" || got.Version != 4 {
		t.Fatal(got, e)
	}
	current = got
	update = objectDomainEntity.Update{Content: got.Content, ExpectedVersion: 4}
	got, e = objectDomainEntity.ApplyUpdate(current, "A", update)
	if e != nil || got.Version != 4 || got.CompletedQuantity != "5.5" {
		t.Fatal("reset/noop", got, e)
	}
	update.CompletedQuantity = ptr("")
	if _, e = objectDomainEntity.ApplyUpdate(current, "A", update); !errors.Is(e, objectDomainEntity.ErrInput) {
		t.Fatal("cleared work", e)
	}
	update.CompletedQuantity = nil
	update.Status = objectDomainEntity.StatusCompleted
	if _, e = objectDomainEntity.ApplyUpdate(current, "A", update); !errors.Is(e, objectDomainEntity.ErrInput) {
		t.Fatal("contradictory status", e)
	}
	update.Status = ""
	update.Type = objectDomainEntity.TypeNonWork
	if _, e = objectDomainEntity.ApplyUpdate(current, "A", update); !errors.Is(e, objectDomainEntity.ErrProgressRemoval) {
		t.Fatal("lost actual work", e)
	}
	update.Type = objectDomainEntity.TypeWork
	update.Quantity = "2"
	if _, e = objectDomainEntity.ApplyUpdate(current, "A", update); !errors.Is(e, objectDomainEntity.ErrInput) {
		t.Fatal("shrank below done", e)
	}
	update.CompletedQuantity = ptr("2")
	got, e = objectDomainEntity.ApplyUpdate(current, "A", update)
	if e != nil || got.Status != objectDomainEntity.StatusCompleted || got.Quantity != "2" || got.CompletedQuantity != "2" {
		t.Fatal("batch adjustment", got, e)
	}
	update.ExpectedVersion = 2
	if _, e = objectDomainEntity.ApplyUpdate(current, "A", update); !errors.Is(e, objectDomainEntity.ErrConflict) {
		t.Fatal(e)
	}
	current.CompletedQuantity = "0"
	current.Status = objectDomainEntity.StatusNotStarted
	update = objectDomainEntity.Update{Content: current.Content, ExpectedVersion: current.Version}
	update.Type = objectDomainEntity.TypeNonWork
	got, e = objectDomainEntity.ApplyUpdate(current, "A", update)
	if e != nil || got.CompletedQuantity != "" || got.Status != objectDomainEntity.StatusNotStarted {
		t.Fatal("zero conversion", got, e)
	}
	update = objectDomainEntity.Update{Content: got.Content, ExpectedVersion: got.Version}
	update.Type = objectDomainEntity.TypeWork
	if _, e = objectDomainEntity.ApplyUpdate(got, "A", update); !errors.Is(e, objectDomainEntity.ErrInput) {
		t.Fatal("nonwork conversion reset", e)
	}
}

func TestHistoricalWorkWithoutQuantityRemainsLegacy(t *testing.T) {
	c := content()
	c.Type = objectDomainEntity.TypeWork
	c.Quantity = "11"
	current := objectDomainEntity.Object{Content: c, CreatorID: "A", Status: objectDomainEntity.StatusInProgress, Version: 1}
	got, e := objectDomainEntity.ApplyUpdate(current, "A", objectDomainEntity.Update{Content: c, Status: objectDomainEntity.StatusCompleted, ExpectedVersion: 1})
	if e != nil || got.Status != objectDomainEntity.StatusCompleted || got.CompletedQuantity != "" {
		t.Fatal(got, e)
	}
	got, e = objectDomainEntity.ApplyUpdate(current, "A", objectDomainEntity.Update{Content: c, CompletedQuantity: ptr("4"), ExpectedVersion: 1})
	if e != nil || got.Status != objectDomainEntity.StatusInProgress || got.CompletedQuantity != "4" {
		t.Fatal(got, e)
	}
}

func TestLegacyNonWorkRenamePreservesTechnicalStatus(t *testing.T) {
	c := content()
	c.Type = objectDomainEntity.TypeNonWork
	current := objectDomainEntity.Object{Content: c, CreatorID: "A", Status: objectDomainEntity.StatusCompleted, Version: 3}
	update := objectDomainEntity.Update{Content: c, ExpectedVersion: 3}
	update.Name = "renamed"
	got, err := objectDomainEntity.ApplyUpdate(current, "A", update)
	if err != nil || got.Status != objectDomainEntity.StatusCompleted || got.CompletedQuantity != "" || got.Name != "renamed" {
		t.Fatalf("legacy non-work rename got %+v, %v", got, err)
	}
	update = objectDomainEntity.Update{Content: got.Content, Status: objectDomainEntity.StatusInProgress, ExpectedVersion: got.Version}
	got, err = objectDomainEntity.ApplyUpdate(got, "A", update)
	if err != nil || got.Status != objectDomainEntity.StatusInProgress {
		t.Fatalf("legacy explicit status got %+v, %v", got, err)
	}
}
