package main

import (
	"math"
	"regexp"
	"strings"
	"testing"
)

func TestLocationNumberInRange(t *testing.T) {
	for _, tc := range []struct {
		name  string
		value float64
		valid bool
	}{
		{"minimum", -90, true},
		{"maximum", 90, true},
		{"zero", 0, true},
		{"below", -91, false},
		{"above", 91, false},
		{"NaN", math.NaN(), false},
		{"positive infinity", math.Inf(1), false},
		{"negative infinity", math.Inf(-1), false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			if got := locationNumberInRange(tc.value, -90, 90); got != tc.valid {
				t.Fatalf("valid=%v want=%v", got, tc.valid)
			}
		})
	}
}

func TestRequestIDPattern(t *testing.T) {
	for _, tc := range []struct {
		name, id string
		valid    bool
	}{
		{"hex", strings.Repeat("0123456789abcdef", 4), true},
		{"empty", "", false},
		{"short", strings.Repeat("a", 63), false},
		{"long", strings.Repeat("a", 65), false},
		{"uppercase", strings.Repeat("A", 64), false},
		{"nonhex", strings.Repeat("g", 64), false},
		{"newline", strings.Repeat("a", 64) + "\n", false},
		{"unicode", strings.Repeat("ａ", 64), false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			if got := requestIDPattern.MatchString(tc.id); got != tc.valid {
				t.Fatalf("valid=%v want=%v", got, tc.valid)
			}
		})
	}
}

func BenchmarkRequestIDValidation(b *testing.B) {
	id := strings.Repeat("0123456789abcdef", 4)
	b.Run("compile_per_request", func(b *testing.B) {
		b.ReportAllocs()
		for b.Loop() {
			if !regexp.MustCompile(`^[0-9a-f]{64}$`).MatchString(id) {
				b.Fatal("valid ID rejected")
			}
		}
	})
	b.Run("compiled_once", func(b *testing.B) {
		b.ReportAllocs()
		for b.Loop() {
			if !requestIDPattern.MatchString(id) {
				b.Fatal("valid ID rejected")
			}
		}
	})
}
