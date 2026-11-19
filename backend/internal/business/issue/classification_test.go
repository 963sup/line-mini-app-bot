package issue

import (
	"errors"
	"strings"
	"testing"
)

func TestMemberClassificationCannotEditContent(t *testing.T) {
	current := Issue{CreatorID: "author", Title: "Original", Version: 1, State: "open"}
	ids := []string{strings.Repeat("a", 32)}
	change := Update{Content: Content{Title: current.Title}, LabelIDs: &ids, ExpectedVersion: 1, State: "open"}
	result, err := ApplyUpdate(current, "member", change)
	if err != nil || len(result.LabelIDs) != 1 || result.Version != 2 {
		t.Fatalf("member classification: %+v %v", result, err)
	}
	change.Body = "unauthorized edit"
	if _, err := ApplyUpdate(current, "member", change); !errors.Is(err, ErrForbidden) {
		t.Fatalf("mixed edit: %v", err)
	}
}
