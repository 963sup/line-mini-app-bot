package object

import (
	"errors"
	"strings"
	"testing"
)

func TestSpaceBindingOptionalAndUpdates(t *testing.T) {
	c := content()
	c.SpaceID = "C" + strings.Repeat("a", 32)
	if _, err := Normalize(c); !errors.Is(err, ErrInput) {
		t.Fatalf("LINE group accepted as business space: %v", err)
	}
	c.SpaceID = strings.Repeat("a", 32)
	current := Object{Content: c, CreatorID: "A", Status: StatusNotStarted, Version: 1}
	u := Update{Content: content(), ExpectedVersion: 1}
	got, err := ApplyUpdate(current, "A", u)
	if err != nil || got.SpaceID != c.SpaceID {
		t.Fatalf("omission lost binding: %+v %v", got, err)
	}
	empty := ""
	u.SpaceID = &empty
	got, err = ApplyUpdate(current, "A", u)
	if err != nil || got.SpaceID != "" || got.Version != 2 {
		t.Fatalf("explicit clear: %+v %v", got, err)
	}
	next := strings.Repeat("b", 32)
	u.SpaceID = &next
	got, err = ApplyUpdate(current, "A", u)
	if err != nil || got.SpaceID != next {
		t.Fatalf("explicit assignment: %+v %v", got, err)
	}
}
