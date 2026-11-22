package command

import (
	"context"
	"strings"
	"time"

	attendanceport "github.com/963sup/line-mini-app-bot/internal/business/attendance/application/port"
	attendanceentity "github.com/963sup/line-mini-app-bot/internal/business/attendance/domain/entity"
	attendancepolicy "github.com/963sup/line-mini-app-bot/internal/business/attendance/domain/policy"
	attendanceservice "github.com/963sup/line-mini-app-bot/internal/business/attendance/domain/service"
)

func ApproveCorrection(ctx context.Context, runner attendanceport.CorrectionReviewRunner, actor, id string, now time.Time) error {
	return runCorrectionReview(ctx, runner, actor, id, attendancepolicy.Decision{Status: "approved"}, now)
}

func RejectCorrection(ctx context.Context, runner attendanceport.CorrectionReviewRunner, actor, id, reason string, now time.Time) error {
	reason = strings.TrimSpace(reason)
	if reason == "" || len([]rune(reason)) > 500 {
		return attendanceentity.ErrCorrection
	}
	return runCorrectionReview(ctx, runner, actor, id, attendancepolicy.Decision{Status: "rejected", Reason: reason}, now)
}

func CancelCorrection(ctx context.Context, runner attendanceport.CorrectionReviewRunner, actor, id string, now time.Time) error {
	return runCorrectionReview(ctx, runner, actor, id, attendancepolicy.Decision{Status: "cancelled"}, now)
}

func runCorrectionReview(ctx context.Context, runner attendanceport.CorrectionReviewRunner, actor, id string, decision attendancepolicy.Decision, now time.Time) error {
	return runner.RunCorrectionReview(ctx, actor, id, func(tx attendanceport.ReviewTransaction, actorKey string, correction attendanceentity.Correction) error {
		return ReviewCorrection(tx, actorKey, correction, decision, now)
	})
}

func ReviewCorrection(tx attendanceport.ReviewTransaction, actorKey string, c attendanceentity.Correction, d attendancepolicy.Decision, now time.Time) error {
	switch d.Status {
	case "cancelled":
		if actorKey != c.UserKey {
			return attendanceentity.ErrCorrection
		}
	case "approved", "rejected":
		groups, err := tx.ReviewerGroups()
		if err != nil {
			return err
		}
		if !attendancepolicy.CanReview(groups, actorKey, c.UserKey, c.GroupID) {
			return attendanceentity.ErrCorrection
		}
	default:
		return attendanceentity.ErrCorrection
	}
	prior, audit, err := tx.PreviousDecision()
	if err != nil {
		return err
	}
	needed, err := prior.NeedsDecision(d, audit)
	if err != nil || !needed {
		return err
	}
	if d.Status == "approved" {
		before, err := tx.Original(c)
		if err != nil {
			return err
		}
		if c.End.After(now) {
			return attendanceentity.ErrCorrection
		}
		after, err := attendanceservice.ApplyCorrection(before, c)
		if err != nil {
			return err
		}
		if err := tx.SaveApproval(c, before, after, actorKey, now); err != nil {
			return err
		}
	}
	return tx.SaveDecision(d, actorKey, now)
}
