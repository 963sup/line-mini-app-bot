package navigation

import "testing"

func TestRegistrationRejectsConflicts(t *testing.T) {
	first := route{command: "/menu", aliases: []string{"選單"}, action: "menu"}
	for name, second := range map[string]route{
		"command":  {command: " /MENU "},
		"alias":    {command: "/help", aliases: []string{"選單"}},
		"postback": {command: "/help", action: "menu"},
	} {
		t.Run(name, func(t *testing.T) {
			defer func() {
				if recover() == nil {
					t.Fatal("conflicting registration accepted")
				}
			}()
			indexRoutes(first, second)
		})
	}
}
