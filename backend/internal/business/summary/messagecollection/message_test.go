package messagecollection

import (
	"testing"
	"time"
)

func TestCollectionWindow(t *testing.T) {
	now := time.Unix(1700000000, 0)
	for _, tc := range []struct {
		offset time.Duration
		want   bool
	}{
		{-24*time.Hour - time.Nanosecond, false},
		{-24 * time.Hour, true},
		{time.Minute, true},
		{time.Minute + time.Nanosecond, false},
	} {
		m := Message{ID: "m", UserID: "u", OccurredAt: now.Add(tc.offset)}
		if m.CanCollect(now) != tc.want {
			t.Errorf("offset %v: want %t", tc.offset, tc.want)
		}
	}
}

func TestEditPreservesOriginalExpiryAndRejectsReplay(t *testing.T) {
	at := time.Unix(1700000000, 0)
	m := Message{OccurredAt: at}
	m.Retain(3)
	if !m.ApplyEdit("first", at.UnixMilli(), 3, at) {
		t.Fatal("edit at original timestamp rejected")
	}
	if m.ApplyEdit("replay", at.UnixMilli(), 3, at) ||
		m.ApplyEdit("old epoch", at.UnixMilli()+1, 2, at) ||
		m.ApplyEdit("expired", at.UnixMilli()+1, 3, at.Add(24*time.Hour)) {
		t.Fatal("invalid edit accepted")
	}
	if m.Text != "first" || !m.ExpireAt.Equal(at.Add(24*time.Hour)) {
		t.Fatal("rejected edit changed retained message")
	}
}
