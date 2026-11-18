package object

import (
	"errors"
	"testing"
)

func ptr(s string) *string { return &s }
func TestExactWorkStatus(t *testing.T) {
	for _, c := range []struct{ total, done, status string }{{"11", "0", StatusNotStarted}, {"11", "5.5", StatusInProgress}, {"11", "11.000000", StatusCompleted}, {"999999999999999999.999999", "999999999999999999.999998", StatusInProgress}, {"0.000001", "0.000001", StatusCompleted}} {
		got, e := WorkStatus(c.total, c.done)
		if e != nil || got != c.status {
			t.Fatal(c, got, e)
		}
	}
	for _, done := range []string{"-1", "12", "NaN", "Inf", "1e1", "1/2", "0.0000001", "1000000000000000000"} {
		if _, e := WorkStatus("11", done); !errors.Is(e, ErrInput) {
			t.Fatal(done, e)
		}
	}
}

func TestProgressConversionsAndAbsoluteUpdates(t *testing.T) {
	base := content()
	base.Quantity = "11"
	base.Type = "Engineering"
	current := Object{Content: base, CreatorID: "A", Status: StatusInProgress, Version: 3}
	update := Update{Content: base, ExpectedVersion: 3}
	got, e := ApplyUpdate(current, "A", update)
	if e != nil || got.Version != 3 || got.Status != StatusInProgress {
		t.Fatal("legacy noop", got, e)
	}
	update.Status = StatusCompleted
	got, e = ApplyUpdate(current, "A", update)
	if e != nil || got.Status != StatusCompleted || got.CompletedQuantity != "" {
		t.Fatal("legacy status", got, e)
	}
	update.Type = " work "
	update.Status = ""
	if _, e = ApplyUpdate(current, "A", update); !errors.Is(e, ErrInput) {
		t.Fatal("conversion silently defaulted", e)
	}
	update.CompletedQuantity = ptr("5.500000")
	got, e = ApplyUpdate(current, "A", update)
	if e != nil || got.Status != StatusInProgress || got.CompletedQuantity != "5.5" || got.Version != 4 {
		t.Fatal(got, e)
	}
	current = got
	update = Update{Content: got.Content, ExpectedVersion: 4}
	got, e = ApplyUpdate(current, "A", update)
	if e != nil || got.Version != 4 || got.CompletedQuantity != "5.5" {
		t.Fatal("reset/noop", got, e)
	}
	update.CompletedQuantity = ptr("")
	if _, e = ApplyUpdate(current, "A", update); !errors.Is(e, ErrInput) {
		t.Fatal("cleared work", e)
	}
	update.CompletedQuantity = nil
	update.Status = StatusCompleted
	if _, e = ApplyUpdate(current, "A", update); !errors.Is(e, ErrInput) {
		t.Fatal("contradictory status", e)
	}
	update.Status = ""
	update.Type = TypeNonWork
	if _, e = ApplyUpdate(current, "A", update); !errors.Is(e, ErrProgressRemoval) {
		t.Fatal("lost actual work", e)
	}
	update.Type = TypeWork
	update.Quantity = "2"
	if _, e = ApplyUpdate(current, "A", update); !errors.Is(e, ErrInput) {
		t.Fatal("shrank below done", e)
	}
	update.CompletedQuantity = ptr("2")
	got, e = ApplyUpdate(current, "A", update)
	if e != nil || got.Status != StatusCompleted || got.Quantity != "2" || got.CompletedQuantity != "2" {
		t.Fatal("batch adjustment", got, e)
	}
	update.ExpectedVersion = 2
	if _, e = ApplyUpdate(current, "A", update); !errors.Is(e, ErrConflict) {
		t.Fatal(e)
	}
	current.CompletedQuantity = "0"
	current.Status = StatusNotStarted
	update = Update{Content: current.Content, ExpectedVersion: current.Version}
	update.Type = TypeNonWork
	got, e = ApplyUpdate(current, "A", update)
	if e != nil || got.CompletedQuantity != "" || got.Status != StatusNotStarted {
		t.Fatal("zero conversion", got, e)
	}
	update = Update{Content: got.Content, ExpectedVersion: got.Version}
	update.Type = TypeWork
	if _, e = ApplyUpdate(got, "A", update); !errors.Is(e, ErrInput) {
		t.Fatal("nonwork conversion reset", e)
	}
}

func TestHistoricalWorkWithoutQuantityRemainsLegacy(t *testing.T) {
	c := content()
	c.Type = TypeWork
	c.Quantity = "11"
	current := Object{Content: c, CreatorID: "A", Status: StatusInProgress, Version: 1}
	got, e := ApplyUpdate(current, "A", Update{Content: c, Status: StatusCompleted, ExpectedVersion: 1})
	if e != nil || got.Status != StatusCompleted || got.CompletedQuantity != "" {
		t.Fatal(got, e)
	}
	got, e = ApplyUpdate(current, "A", Update{Content: c, CompletedQuantity: ptr("4"), ExpectedVersion: 1})
	if e != nil || got.Status != StatusInProgress || got.CompletedQuantity != "4" {
		t.Fatal(got, e)
	}
}

func TestLegacyNonWorkRenamePreservesTechnicalStatus(t *testing.T) {
	c := content()
	c.Type = TypeNonWork
	current := Object{Content: c, CreatorID: "A", Status: StatusCompleted, Version: 3}
	update := Update{Content: c, ExpectedVersion: 3}
	update.Name = "renamed"
	got, err := ApplyUpdate(current, "A", update)
	if err != nil || got.Status != StatusCompleted || got.CompletedQuantity != "" || got.Name != "renamed" {
		t.Fatalf("legacy non-work rename got %+v, %v", got, err)
	}
	update = Update{Content: got.Content, Status: StatusInProgress, ExpectedVersion: got.Version}
	got, err = ApplyUpdate(got, "A", update)
	if err != nil || got.Status != StatusInProgress {
		t.Fatalf("legacy explicit status got %+v, %v", got, err)
	}
}
