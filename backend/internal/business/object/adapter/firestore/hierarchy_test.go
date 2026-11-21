package firestore

import (
	"errors"
	"strings"
	"testing"

	entity "github.com/963sup/line-mini-app-bot/internal/business/object/domain/entity"
)

func TestEmulatorObjectHierarchy(t *testing.T) {
	ctx, store, repo := emulatorStore(t)
	const actor = "actor-a"
	content := testContent()
	root, err := store.Create(ctx, repo, actor, "hierarchy-root", content)
	if err != nil {
		t.Fatal(err)
	}
	content.ParentID = root.ID
	child, err := store.Create(ctx, repo, actor, "hierarchy-child", content)
	if err != nil || child.ParentID != root.ID {
		t.Fatalf("child parent = %q, error = %v", child.ParentID, err)
	}
	content.ParentID = child.ID
	grandchild, err := store.Create(ctx, repo, actor, "hierarchy-grandchild", content)
	if err != nil || grandchild.ParentID != child.ID {
		t.Fatalf("grandchild parent = %q, error = %v", grandchild.ParentID, err)
	}
	if _, err := store.Create(ctx, repo, actor, "hierarchy-child", content); !errors.Is(err, entity.ErrCreationConflict) {
		t.Fatalf("changing parent on retry returned %v", err)
	}
	content.ParentID = strings.Repeat("f", 32)
	if _, err := store.Create(ctx, repo, actor, "hierarchy-missing", content); !errors.Is(err, entity.ErrNotFound) {
		t.Fatalf("missing parent returned %v", err)
	}
}
