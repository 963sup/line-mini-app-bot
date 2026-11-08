package firestore

import (
	"context"
	"time"

	db "cloud.google.com/go/firestore"
	"github.com/963sup/line-mini-app-bot/internal/attendance"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// ProposeCorrection is always scoped to the authenticated employee.
func (s *Attendance) ProposeCorrection(ctx context.Context, user string, c attendance.Correction, now time.Time) (string, error) {
	c.UserKey = documentKey(user)
	c.CreatedAt = now
	ref := s.Client.Collection("channels").Doc(s.ChannelID).Collection("attendanceCorrectionRequests").NewDoc()
	c.ID = ref.ID
	err := s.Client.RunTransaction(ctx, func(ctx context.Context, tx *db.Transaction) error {
		snapshot, err := tx.Get(s.user(user).Collection("attendance").Doc(c.RecordID))
		if err != nil {
			return err
		}
		var before attendance.Record
		if err := snapshot.DataTo(&before); err != nil {
			return err
		}
		before.RecordID = c.RecordID
		if c.End.After(now) {
			return attendance.ErrCorrection
		}
		if _, err := attendance.ApplyCorrection(before, c); err != nil {
			return err
		}
		return tx.Create(ref, c)
	})
	return c.ID, err
}

// ApproveCorrection requires a trusted, pre-provisioned group reviewer and
// writes an immutable before/after audit document alongside the new projection.
func (s *Attendance) ApproveCorrection(ctx context.Context, reviewer, id string, now time.Time) error {
	channel := s.Client.Collection("channels").Doc(s.ChannelID)
	return s.Client.RunTransaction(ctx, func(ctx context.Context, tx *db.Transaction) error {
		role, err := tx.Get(channel.Collection("attendanceReviewers").Doc(documentKey(reviewer)))
		if err != nil {
			return attendance.ErrCorrection
		}
		var access struct {
			Groups []string `firestore:"groups"`
		}
		if err := role.DataTo(&access); err != nil {
			return err
		}
		proposal, err := tx.Get(channel.Collection("attendanceCorrectionRequests").Doc(id))
		if err != nil {
			return err
		}
		var c attendance.Correction
		if err := proposal.DataTo(&c); err != nil {
			return err
		}
		allowed := false
		for _, g := range access.Groups {
			if g == c.GroupID {
				allowed = true
			}
		}
		if !allowed || c.UserKey == documentKey(reviewer) {
			return attendance.ErrCorrection
		}
		audit := channel.Collection("attendanceCorrections").Doc(id)
		prior, err := tx.Get(audit)
		if err == nil && prior.Exists() {
			return nil
		}
		if err != nil && !isMissingCorrection(err) {
			return err
		}
		employee := channel.Collection("users").Doc(c.UserKey)
		ref := employee.Collection("attendance").Doc(c.RecordID)
		snapshot, err := tx.Get(ref)
		if err != nil {
			return err
		}
		var before attendance.Record
		if err := snapshot.DataTo(&before); err != nil {
			return err
		}
		before.RecordID = c.RecordID
		after, err := attendance.ApplyCorrection(before, c)
		if err != nil {
			return err
		}
		state := employee.Collection("attendanceStates").Doc(documentKey(c.GroupID))
		activeSnapshot, err := tx.Get(state)
		var active struct {
			RecordID string `firestore:"recordID"`
		}
		if err == nil {
			if err := activeSnapshot.DataTo(&active); err != nil {
				return err
			}
		} else if !isMissingCorrection(err) {
			return err
		}
		if active.RecordID == c.RecordID {
			if err := tx.Update(state, []db.Update{{Path: "recordID", Value: ""}}); err != nil {
				return err
			}
		}
		if err := tx.Set(ref, after); err != nil {
			return err
		}
		return tx.Create(audit, map[string]any{"proposal": c, "before": before, "after": after, "reviewerKey": documentKey(reviewer), "approvedAt": now})
	})
}

func isMissingCorrection(err error) bool { return status.Code(err) == codes.NotFound }

func (s *Attendance) CorrectionStatus(ctx context.Context, user, id string) (string, error) {
	channel := s.Client.Collection("channels").Doc(s.ChannelID)
	result := "pending"
	err := s.Client.RunTransaction(ctx, func(ctx context.Context, tx *db.Transaction) error {
		result = "pending"
		snapshot, err := tx.Get(channel.Collection("attendanceCorrectionRequests").Doc(id))
		if err != nil {
			return err
		}
		var c attendance.Correction
		if err := snapshot.DataTo(&c); err != nil {
			return err
		}
		if c.UserKey != documentKey(user) {
			return attendance.ErrCorrection
		}
		auditSnapshot, auditErr := tx.Get(channel.Collection("attendanceCorrections").Doc(id))
		err = auditErr
		if err == nil {
			audit, decodeErr := decodeCorrectionAudit(auditSnapshot, user, id)
			if decodeErr != nil {
				return decodeErr
			}
			result = "approved"
			if !attendance.VerifyCorrection(audit.Before, audit.Proposal, audit.After, audit.After) {
				result = "invalid"
			}
			return nil
		}
		if !isMissingCorrection(err) {
			return err
		}
		record, err := tx.Get(s.user(user).Collection("attendance").Doc(c.RecordID))
		if err != nil {
			return err
		}
		var current attendance.Record
		if err := record.DataTo(&current); err != nil {
			return err
		}
		if current.Revision != c.ExpectedRevision || current.EventSequence != c.ExpectedEventSequence {
			result = "stale"
		}
		return nil
	})
	return result, err
}

func (s *Attendance) verifyCorrection(ctx context.Context, user string, current attendance.Record) (bool, error) {
	if current.CorrectionID == "" {
		return false, nil
	}
	snapshot, err := s.Client.Collection("channels").Doc(s.ChannelID).Collection("attendanceCorrections").Doc(current.CorrectionID).Get(ctx)
	if isMissingCorrection(err) {
		return false, nil
	}
	if err != nil {
		return false, err
	}
	audit, err := decodeCorrectionAudit(snapshot, user, current.CorrectionID)
	if err != nil {
		if err == attendance.ErrCorrection {
			return false, nil
		}
		return false, err
	}
	return attendance.VerifyCorrection(audit.Before, audit.Proposal, audit.After, current), nil
}

type correctionAudit struct {
	Proposal    attendance.Correction `firestore:"proposal"`
	Before      attendance.Record     `firestore:"before"`
	After       attendance.Record     `firestore:"after"`
	ReviewerKey string                `firestore:"reviewerKey"`
	ApprovedAt  time.Time             `firestore:"approvedAt"`
}

func decodeCorrectionAudit(snapshot *db.DocumentSnapshot, user, id string) (correctionAudit, error) {
	var audit correctionAudit
	if err := snapshot.DataTo(&audit); err != nil {
		return audit, err
	}
	if audit.Proposal.UserKey != documentKey(user) || audit.Proposal.ID != id || audit.ReviewerKey == "" || audit.ReviewerKey == audit.Proposal.UserKey || audit.ApprovedAt.IsZero() {
		return audit, attendance.ErrCorrection
	}
	return audit, nil
}

// CorrectionReview exposes proposal details only to a group-authorized reviewer.
func (s *Attendance) CorrectionReview(ctx context.Context, reviewer, id string) (attendance.Correction, error) {
	channel := s.Client.Collection("channels").Doc(s.ChannelID)
	var c attendance.Correction
	role, err := channel.Collection("attendanceReviewers").Doc(documentKey(reviewer)).Get(ctx)
	if err != nil {
		return c, attendance.ErrCorrection
	}
	var access struct {
		Groups []string `firestore:"groups"`
	}
	if err := role.DataTo(&access); err != nil {
		return c, err
	}
	snapshot, err := channel.Collection("attendanceCorrectionRequests").Doc(id).Get(ctx)
	if err != nil {
		return c, err
	}
	if err := snapshot.DataTo(&c); err != nil {
		return c, err
	}
	allowed := false
	for _, g := range access.Groups {
		if g == c.GroupID {
			allowed = true
		}
	}
	if !allowed || c.UserKey == documentKey(reviewer) {
		return attendance.Correction{}, attendance.ErrCorrection
	}
	return c, nil
}
