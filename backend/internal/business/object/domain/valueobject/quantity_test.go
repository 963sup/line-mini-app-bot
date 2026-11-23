package valueobject

import "testing"

func TestQuantity(t *testing.T) {
	for _, s := range []string{"", "01", "-1", "1e2", "1234567890123456789", "1.0000001"} {
		if _, ok := ParseQuantity(s); ok {
			t.Fatal(s)
		}
	}
	for _, tc := range []struct {
		a, b string
		cmp  int
	}{{"0.000", "0", 0}, {"1.20", "1.2", 0}, {"1.001", "1.01", -1}, {"10", "9.99", 1}, {"999999999999999999.999999", "999999999999999999.999998", 1}} {
		a, ok := ParseQuantity(tc.a)
		b, valid := ParseQuantity(tc.b)
		if !ok || !valid || a.Compare(b) != tc.cmp {
			t.Fatal(tc)
		}
		if tc.cmp == 0 && a != b {
			t.Fatal(tc)
		}
	}
}
