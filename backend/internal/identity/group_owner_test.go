package identity

import (
	"strings"
	"testing"
)

func TestIsGroupOwnerID(t *testing.T) {
	suffix := strings.Repeat("af", 16)
	if !IsGroupOwnerID("C" + suffix) {
		t.Fatal("valid group rejected")
	}
	for _, id := range []string{"", "C", "U" + suffix, "C" + strings.ToUpper(suffix), "C" + suffix[:31], "C" + suffix + "0"} {
		if IsGroupOwnerID(id) {
			t.Errorf("accepted %q", id)
		}
	}
}
