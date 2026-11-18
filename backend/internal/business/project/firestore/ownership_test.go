package firestore

import (
	"strings"
	"testing"

	project "github.com/963sup/line-mini-app-bot/internal/business/project"
)

func TestProjectOwnershipIndependentOfSelectedRepository(t *testing.T) {
	owner := "C" + strings.Repeat("a", 32)
	defaultRepo, otherRepo := strings.Repeat("b", 32), strings.Repeat("c", 32)
	doc := encodeProject(project.Project{OwnerID: owner, DefaultRepositoryID: defaultRepo})
	if !doc.belongsTo(project.Scope{OwnerID: owner, SourceRepositoryID: otherRepo}) {
		t.Fatal("selecting another repository changed ownership")
	}
	if doc.belongsTo(project.Scope{OwnerID: "C" + otherRepo, SourceRepositoryID: defaultRepo}) {
		t.Fatal("default repository granted access across Groups")
	}
	if doc.belongsTo(project.Scope{Legacy: true, OwnerID: owner, SourceRepositoryID: otherRepo}) {
		t.Fatal("repository-scoped endpoint returned another repository's project")
	}
}
