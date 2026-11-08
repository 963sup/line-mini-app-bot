package summary

import (
	"strings"
	"testing"
)

func TestValidateSummary(t *testing.T) {
	for _, tc := range []struct {
		name, text string
		sources    int
		valid      bool
	}{
		{"valid", "  決定 [S1]、待辦 [S2]  ", 2, true},
		{"repeated citation", "決定 [S1] [S1]", 1, true},
		{"empty", "  ", 1, false},
		{"missing citation", "決定", 1, false},
		{"zero citation", "決定 [S0]", 1, false},
		{"unknown citation", "決定 [S2]", 1, false},
		{"overflow citation", "[S999999999999999999999999999999]", 1, false},
		{"length boundary", strings.Repeat("文", 3996) + "[S1]", 1, true},
		{"too long", strings.Repeat("文", 3997) + "[S1]", 1, false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			got, err := validateSummary(tc.text, tc.sources)
			if (err == nil) != tc.valid {
				t.Fatalf("valid=%v error=%v", tc.valid, err)
			}
			if tc.valid && got != strings.TrimSpace(tc.text) {
				t.Fatal("validation changed summary content")
			}
		})
	}
}

func BenchmarkValidateSummary(b *testing.B) {
	text := strings.Repeat("討論重點與決定 [S1]；待辦 [S2]。", 20)
	b.ReportAllocs()
	for b.Loop() {
		if _, err := validateSummary(text, 2); err != nil {
			b.Fatal(err)
		}
	}
}
