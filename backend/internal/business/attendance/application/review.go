package application

import (
	"time"

	"github.com/963sup/line-mini-app-bot/internal/business/attendance/domain/approvals"
	"github.com/963sup/line-mini-app-bot/internal/business/attendance/domain/corrections"
	"github.com/963sup/line-mini-app-bot/internal/business/attendance/domain/records"
)

// ReviewTransaction supplies one consistent view and commits both writes atomically.
type ReviewTransaction interface {
	ReviewerGroups() ([]string, error)
	PreviousDecision() (approvals.Decision, bool, error)
	Original(corrections.Correction) (records.Record, error)
	SaveApproval(corrections.Correction, records.Record, records.Record, string, time.Time) error
	SaveDecision(approvals.Decision, string, time.Time) error
}

func ReviewCorrection(tx ReviewTransaction, actorKey string, c corrections.Correction, d approvals.Decision, now time.Time) error {
	switch d.Status {
	case "cancelled":
		if actorKey != c.UserKey {
			return corrections.ErrCorrection
		}
	case "approved", "rejected":
		groups, err := tx.ReviewerGroups()
		if err != nil {
			return err
		}
		if !approvals.CanReview(groups, actorKey, c.UserKey, c.GroupID) {
			return corrections.ErrCorrection
		}
	default:
		return corrections.ErrCorrection
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
			return corrections.ErrCorrection
		}
		after, err := corrections.ApplyCorrection(before, c)
		if err != nil {
			return err
		}
		if err := tx.SaveApproval(c, before, after, actorKey, now); err != nil {
			return err
		}
	}
	return tx.SaveDecision(d, actorKey, now)
}
