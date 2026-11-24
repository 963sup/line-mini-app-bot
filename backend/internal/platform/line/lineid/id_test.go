package lineid

import (
	"strings"
	"testing"
)

func TestValidGroup(t *testing.T) {
	suffix := strings.Repeat("af", 16)
	if !ValidGroup("C" + suffix) {
		t.Fatal("valid group rejected")
	}
	for _, id := range []string{"", "C", "U" + suffix, "C" + strings.ToUpper(suffix), "C" + suffix[:31], "C" + suffix + "0"} {
		if ValidGroup(id) {
			t.Errorf("accepted %q", id)
		}
	}
}
