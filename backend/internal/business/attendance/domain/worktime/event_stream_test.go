package worktime

import (
	"testing"
	"time"
)

func TestValidateEventStream(t *testing.T) {
	base := []Event{{ID: "in", Kind: "clock-in", At: time.Now(), Sequence: 1, ResultCode: "clocked_in", RecordID: "r", GroupID: "g", ContractVersion: 1}, {ID: "confirm", Kind: "confirm-rest", At: time.Now(), Sequence: 2, ResultCode: "rest_confirmed", RecordID: "r", GroupID: "g", ContractVersion: 1}}
	if issues := ValidateEventStream(base, "r", "g", 2); len(issues) != 0 {
		t.Fatal(issues)
	}
	for _, mutate := range []func([]Event){func(e []Event) { e[1].Sequence = 3 }, func(e []Event) { e[1].ContractVersion = 2 }, func(e []Event) { e[1].GroupID = "other" }, func(e []Event) { e[1].ID = e[0].ID }} {
		input := append([]Event(nil), base...)
		mutate(input)
		if len(ValidateEventStream(input, "r", "g", 2)) == 0 {
			t.Fatal("invalid stream accepted")
		}
	}
}
