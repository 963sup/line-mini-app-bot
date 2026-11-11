package worktime

import "sort"

const EventContractVersion = 1

// ValidateEventStream checks a single shift, before any replay or corrections.
func ValidateEventStream(events []Event, recordID, groupID string, expected int64) []Issue {
	fail := func(code string) []Issue { return []Issue{{SessionID: recordID, Code: code}} }
	if expected < 1 || int64(len(events)) != expected {
		return fail("event_count_mismatch")
	}
	ordered := append([]Event(nil), events...)
	sort.Slice(ordered, func(i, j int) bool { return ordered[i].Sequence < ordered[j].Sequence })
	ids := map[string]bool{}
	for i, e := range ordered {
		if e.ContractVersion != EventContractVersion {
			return fail("unsupported_event_version")
		}
		if e.RecordID != recordID || e.GroupID != groupID {
			return fail("event_scope_mismatch")
		}
		if e.Sequence != int64(i+1) {
			return fail("event_sequence_gap")
		}
		if e.ID == "" || ids[e.ID] || e.At.IsZero() || !AcceptedCode(e.ResultCode) {
			return fail("invalid_event")
		}
		if i == 0 && e.Kind != "clock-in" {
			return fail("missing_initial_clock_in")
		}
		ids[e.ID] = true
	}
	return nil
}
