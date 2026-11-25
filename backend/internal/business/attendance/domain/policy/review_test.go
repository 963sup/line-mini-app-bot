package policy

import (
	"testing"

	attendanceentity "github.com/963sup/line-mini-app-bot/internal/business/attendance/domain/entity"
)

func TestCanReview(t *testing.T) {
	for _, tc := range []struct {
		name                    string
		groups                  []string
		actor, applicant, group string
		want                    bool
	}{
		{"authorized independent reviewer", []string{"group-a"}, "reviewer", "employee", "group-a", true},
		{"self review", []string{"group-a"}, "employee", "employee", "group-a", false},
		{"different group", []string{"group-b"}, "reviewer", "employee", "group-a", false},
		{"no reviewer scope", nil, "reviewer", "employee", "group-a", false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			if got := CanReview(tc.groups, tc.actor, tc.applicant, tc.group); got != tc.want {
				t.Fatalf("CanReview() = %v, want %v", got, tc.want)
			}
		})
	}
}

func TestDecisionRetries(t *testing.T) {
	rejected := Decision{"rejected", "missing evidence"}
	approved := Decision{Status: "approved"}
	for _, tc := range []struct {
		name           string
		prior, request Decision
		audit, needed  bool
		err            error
	}{
		{"first decision", Decision{}, approved, false, true, nil},
		{"same rejection", rejected, rejected, false, false, nil},
		{"changed reason", rejected, Decision{"rejected", "different"}, false, false, attendanceentity.ErrCorrection},
		{"existing approval audit", Decision{}, approved, true, false, nil},
		{"cancel after approval", Decision{}, Decision{Status: "cancelled"}, true, false, attendanceentity.ErrCorrection},
	} {
		t.Run(tc.name, func(t *testing.T) {
			needed, err := tc.prior.NeedsDecision(tc.request, tc.audit)
			if needed != tc.needed || err != tc.err {
				t.Fatalf("needed=%v err=%v", needed, err)
			}
		})
	}
}
