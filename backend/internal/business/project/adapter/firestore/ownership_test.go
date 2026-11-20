package firestore

import (
	projectApplicationPort "github.com/963sup/line-mini-app-bot/internal/business/project/application/port"
	projectDomainEntity "github.com/963sup/line-mini-app-bot/internal/business/project/domain/entity"
)

import (
	"strings"
	"testing"
)

func TestProjectOwnershipIndependentOfSelectedRepository(t *testing.T) {
	owner := "C" + strings.Repeat("a", 32)
	defaultRepo, otherRepo := strings.Repeat("b", 32), strings.Repeat("c", 32)
	doc := encodeProject(projectDomainEntity.Project{OwnerID: owner, DefaultRepositoryID: defaultRepo})
	if !doc.belongsTo(projectApplicationPort.Scope{OwnerID: owner, SourceRepositoryID: otherRepo}) {
		t.Fatal("selecting another repository changed ownership")
	}
	if doc.belongsTo(projectApplicationPort.Scope{OwnerID: "C" + otherRepo, SourceRepositoryID: defaultRepo}) {
		t.Fatal("default repository granted access across Groups")
	}
	if doc.belongsTo(projectApplicationPort.Scope{Legacy: true, OwnerID: owner, SourceRepositoryID: otherRepo}) {
		t.Fatal("repository-scoped endpoint returned another repository's project")
	}
}
