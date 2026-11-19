package issue

import (
	"errors"
	"testing"
)

func TestCloseAndReopenExistingIssues(t *testing.T) {
	for _, state := range []string{"open", "in_progress"} {
		current := Issue{Title: "Repair", CreatorID: "creator", Version: 4, State: state}
		update := Update{Content: Content{Title: current.Title}, ExpectedVersion: current.Version, State: "completed"}
		if _, err := ApplyUpdate(current, "viewer", update); !errors.Is(err, ErrForbidden) {
			t.Fatalf("viewer closing %s: %v", state, err)
		}
		closed, err := ApplyUpdate(current, "creator", update)
		if err != nil || closed.State != "completed" || closed.Version != 5 {
			t.Fatalf("closing %s: %+v %v", state, closed, err)
		}
		update.State, update.ExpectedVersion = "open", closed.Version
		reopened, err := ApplyUpdate(closed, "creator", update)
		if err != nil || reopened.State != "open" || reopened.Version != 6 {
			t.Fatalf("reopening: %+v %v", reopened, err)
		}
	}
}
