package valueobject

import (
	"regexp"
)

var idPattern = regexp.MustCompile(`^C[0-9a-f]{32}$`)

func ValidID(id string) bool { return idPattern.MatchString(id) }
