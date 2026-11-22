package entity

import "testing"

func TestWorkConversionDefault(t *testing.T) {
	for _, tc := range []struct{ kind, status, want string }{
		{"Engineering", StatusNotStarted, "zero"},
		{"Engineering", StatusCompleted, "quantity"},
		{"Engineering", "in_progress", ""},
		{TypeNonWork, StatusCompleted, ""},
		{TypeWork, StatusNotStarted, ""},
	} {
		obj := Object{Content: Content{Type: tc.kind}, Status: tc.status}
		if got := obj.WorkConversionDefault(); got != tc.want {
			t.Errorf("%s/%s: got %q, want %q", tc.kind, tc.status, got, tc.want)
		}
	}
}
