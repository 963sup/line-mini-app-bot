package lineid

import "regexp"

var userPattern = regexp.MustCompile(`^U[a-f0-9]{32}$`)

func ValidUser(id string) bool { return userPattern.MatchString(id) }
