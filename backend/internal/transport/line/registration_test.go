package line

import "testing"

func TestRegistrationAcrossDispatchPhases(t *testing.T) {
	for _, pair := range [][2]Handler{
		{{Owner: "one", Commands: []string{" /Menu "}}, {Owner: "two", Commands: []string{"/menu"}}},
		{{Owner: "one", Postbacks: []string{"open"}}, {Owner: "two", Postbacks: []string{"open"}}},
	} {
		d := Dispatch{EarlyCommands: []Handler{pair[0]}, Commands: []Handler{pair[1]}}
		if err := d.validate(); err == nil {
			t.Fatal("expected cross-adapter registration conflict")
		}
	}
}
