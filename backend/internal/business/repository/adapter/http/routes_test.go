package repositoryhttp

import (
	"strings"
	"testing"

	"github.com/963sup/line-mini-app-bot/internal/business/repository/domain/entity"
)

func TestViewGroupID(t *testing.T) {
	suffix := strings.Repeat("a", 32)
	group := "C" + suffix
	if got := view(entity.Repository{OwnerID: group}).GroupId; got == nil || *got != group {
		t.Fatal("group-owned repository must expose its group ID")
	}
	for _, owner := range []string{"U" + suffix, "", "C-invalid"} {
		if view(entity.Repository{OwnerID: owner}).GroupId != nil {
			t.Errorf("unexpected group ID for owner %q", owner)
		}
	}
}
