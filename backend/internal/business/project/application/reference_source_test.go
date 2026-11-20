package application

import (
	"testing"

	projectApplicationPort "github.com/963sup/line-mini-app-bot/internal/business/project/application/port"

	projectDomainEntity "github.com/963sup/line-mini-app-bot/internal/business/project/domain/entity"
)

func TestReferenceSource(t *testing.T) {
	scope := projectApplicationPort.Scope{OwnerID: "owner", SourceRepositoryID: "home", SelectedSource: projectDomainEntity.Source{ID: "other", OwnerID: "other-owner"}}
	for _, want := range []projectDomainEntity.Source{{ID: "home", OwnerID: "owner"}, scope.SelectedSource} {
		got, err := scope.ReferenceSource(want.ID)
		if err != nil || got != want {
			t.Fatalf("source %q: got %+v, %v", want.ID, got, err)
		}
	}
	if _, err := scope.ReferenceSource("unselected"); err != projectDomainEntity.ErrForbidden {
		t.Fatalf("unselected source: %v", err)
	}
}
