package repositoryhttp

import (
	"strings"
	"testing"

	"github.com/963sup/line-mini-app-bot/internal/business/repository/domain/entity"
)

func TestViewKeepsOwnerIdentityAndInteractionProjection(t *testing.T) {
	owner := "U" + strings.Repeat("a", 32)
	group := "C" + strings.Repeat("b", 32)
	got := view(entity.Repository{OwnerType: entity.OwnerUser, OwnerID: owner, InteractionGroupID: group})
	if got.OwnerId != owner || got.OwnerType != "USER" {
		t.Fatalf("owner = %q %q", got.OwnerType, got.OwnerId)
	}
	if got.GroupId == nil || *got.GroupId != group {
		t.Fatalf("interaction group = %v", got.GroupId)
	}
}
