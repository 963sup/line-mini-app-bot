package valueobject

import (
	"regexp"
	"strings"
)

var decimal = regexp.MustCompile(`^(0|[1-9][0-9]{0,17})(\.[0-9]{1,6})?$`)

// Quantity stores a canonical decimal. Its zero value represents zero.
type Quantity struct{ digits string }

func ParseQuantity(raw string) (Quantity, bool) {
	if !decimal.MatchString(raw) {
		return Quantity{}, false
	}
	if strings.Contains(raw, ".") {
		raw = strings.TrimRight(strings.TrimRight(raw, "0"), ".")
	}
	if raw == "0" {
		return Quantity{}, true
	}
	return Quantity{digits: raw}, true
}

func (q Quantity) String() string {
	if q.digits == "" {
		return "0"
	}
	return q.digits
}
func (q Quantity) IsZero() bool { return q.digits == "" }
func (q Quantity) Compare(other Quantity) int {
	a, af, _ := strings.Cut(q.String(), ".")
	b, bf, _ := strings.Cut(other.String(), ".")
	if len(a) < len(b) {
		return -1
	}
	if len(a) > len(b) {
		return 1
	}
	if cmp := strings.Compare(a, b); cmp != 0 {
		return cmp
	}
	return strings.Compare(af, bf)
}
