package policy

// CanProvision checks whether this entitlement permits the requested creation or retry.
func CanProvision(enabled bool, existing string, allowExisting bool) bool {
	return enabled && (existing == "" || allowExisting)
}
