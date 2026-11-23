package entity

import (
	"errors"
	"reflect"
	"testing"
)

func TestApplyReferenceCheck(t *testing.T) {
	work := Work{Name: "task", Percent: 45, Reference: &Reference{"repository", "r"}, ObjectID: "o", SpaceIDs: []string{"p"}, ObjectName: "object", SpaceNames: []string{"space"}}
	got := work

	if err := got.ApplyReferenceCheck(ErrNotFound, false); err != ErrNotFound || !reflect.DeepEqual(got, work) {
		t.Fatal("write should reject an inaccessible reference without redacting it")
	}
	unavailable := errors.New("storage unavailable")
	if err := got.ApplyReferenceCheck(unavailable, true); err != unavailable {
		t.Fatal("read should propagate a lookup failure")
	}
	if err := got.ApplyReferenceCheck(ErrNotFound, true); err != nil {
		t.Fatal(err)
	}
	want := Work{Name: work.Name, Percent: work.Percent}
	if !reflect.DeepEqual(got, want) {
		t.Fatal("redaction should retain only the journaled work")
	}
}
