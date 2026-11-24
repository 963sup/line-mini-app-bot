package navigation

import (
	"fmt"
	"strings"
)

type route struct {
	command         string
	aliases         []string
	action          string
	privatePostback bool
	textReply       string
}

type routes struct {
	text     map[string]route
	postback map[string]route
}

func indexRoutes(entries ...route) routes {
	r := routes{map[string]route{}, map[string]route{}}
	for _, entry := range entries {
		for _, name := range append([]string{entry.command}, entry.aliases...) {
			key := strings.ToLower(strings.TrimSpace(name))
			if previous, exists := r.text[key]; exists {
				panic(fmt.Sprintf("LINE navigation command %q registered by %q and %q", key, previous.command, entry.command))
			}
			r.text[key] = entry
		}
		if entry.action != "" {
			if previous, exists := r.postback[entry.action]; exists {
				panic(fmt.Sprintf("LINE navigation postback %q registered by %q and %q", entry.action, previous.command, entry.command))
			}
			r.postback[entry.action] = entry
		}
	}
	return r
}

var commonRoutes = []route{
	route{command: "/menu", aliases: []string{"選單", "功能"}},
	route{command: "/help", action: "help"},
	route{command: "/miniapp", action: "miniapp"},
}

// PrivateAction serves a private postback and its existing slash command.
type PrivateAction struct{ Action, Reply string }
