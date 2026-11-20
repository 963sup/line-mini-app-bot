package valueobject

import (
	"regexp"
)

// ValidLeaveKey uses the same canonical UUID shape as other retry-safe submissions.
func ValidLeaveKey(key string) bool { return leaveKeyPattern.MatchString(key) }

var leaveKeyPattern = regexp.MustCompile(`^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$`)
