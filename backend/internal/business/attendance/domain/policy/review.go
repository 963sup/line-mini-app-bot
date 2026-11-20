package policy

import (
	"slices"

	attendanceentity "github.com/963sup/line-mini-app-bot/internal/business/attendance/domain/entity"
)

// CanReview preserves the group-scoped, independent-reviewer rule.
func CanReview(groups []string, actorKey, applicantKey, group string) bool {
	return actorKey != applicantKey && slices.Contains(groups, group)
}

// Decision records the outcome and reason of a correction review.
type Decision struct {
	Status string
	Reason string
}

// NeedsDecision distinguishes a new decision from a successful retry or conflict.
// An approval audit is authoritative even for proposals without a stored status.
func (prior Decision) NeedsDecision(request Decision, hasApproval bool) (bool, error) {
	if hasApproval {
		if request.Status == "approved" {
			return false, nil
		}
		return false, attendanceentity.ErrCorrection
	}
	if prior.Status == "" {
		return true, nil
	}
	if prior == request {
		return false, nil
	}
	return false, attendanceentity.ErrCorrection
}
