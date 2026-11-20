package project

import "testing"

func TestReferenceSource(t *testing.T) {
	scope := Scope{OwnerID: "owner", SourceRepositoryID: "home", SelectedSource: Source{ID: "other", OwnerID: "other-owner"}}
	for _, want := range []Source{{ID: "home", OwnerID: "owner"}, scope.SelectedSource} {
		got, err := scope.ReferenceSource(want.ID)
		if err != nil || got != want {
			t.Fatalf("source %q: got %+v, %v", want.ID, got, err)
		}
	}
	if _, err := scope.ReferenceSource("unselected"); err != ErrForbidden {
		t.Fatalf("unselected source: %v", err)
	}
}
