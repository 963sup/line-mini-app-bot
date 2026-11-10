package lineui

import "strings"

// GroupName displays a known name or a short, non-sensitive group suffix.
func GroupName(group, name string) string {
	if name = strings.TrimSpace(name); name != "" {
		return name
	}
	if len(group) > 6 {
		group = group[len(group)-6:]
	}
	return "群組（" + group + "）"
}
