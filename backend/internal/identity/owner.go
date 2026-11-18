package identity

import "regexp"

var ownerID = regexp.MustCompile(`^[UC][a-f0-9]{32}$`)

func ValidOwner(id string) bool    { return ownerID.MatchString(id) }
func IsUserOwnerID(id string) bool { return ValidOwner(id) && id[0] == 'U' }

func IsGroupOwnerID(id string) bool { return ValidOwner(id) && id[0] == 'C' }
