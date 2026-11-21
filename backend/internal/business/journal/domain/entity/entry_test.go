package entity

import (
	"errors"
	"reflect"
	"testing"
)

func TestCheckReferences(t *testing.T) {
	work := Work{Name: "task", Percent: 45, Reference: &Reference{"repository", "r"}, ObjectID: "o", SpaceIDs: []string{"p"}, ObjectName: "object", SpaceNames: []string{"space"}}
	check := func(Reference) error { return ErrNotFound }
	d := Draft{Today: []Work{work}, Tomorrow: []Work{work}}
	if err := d.CheckReferences(false, check); err != ErrNotFound || !reflect.DeepEqual(d.Today[0], work) {
		t.Fatal("write should reject an inaccessible reference without redacting it")
	}
	unavailable := errors.New("storage unavailable")
	if err := d.CheckReferences(true, func(Reference) error { return unavailable }); err != unavailable {
		t.Fatal("read should propagate a lookup failure")
	}
	if err := d.CheckReferences(true, check); err != nil {
		t.Fatal(err)
	}
	want := Work{Name: work.Name, Percent: work.Percent}
	if !reflect.DeepEqual(d.Today[0], want) || !reflect.DeepEqual(d.Tomorrow[0], want) {
		t.Fatal("both days should retain only the journaled work")
	}
}
