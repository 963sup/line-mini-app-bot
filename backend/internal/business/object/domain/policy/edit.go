package policy

// CanEdit requires an authenticated actor who created the object.
func CanEdit(actor, creator string) bool { return actor != "" && actor == creator }
