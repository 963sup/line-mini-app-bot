package firestore

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"time"

	db "cloud.google.com/go/firestore"
	"github.com/963sup/line-mini-app-bot/internal/attendance"
	"github.com/963sup/line-mini-app-bot/internal/attendance/worktime"
	"github.com/google/uuid"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// ProposeCorrection preserves the existing internal caller contract. New Web
// submissions use ProposeCorrectionOnce so a lost response can be retried safely.
func (s *Attendance) ProposeCorrection(ctx context.Context, user string, c attendance.Correction, now time.Time) (_ string, err error) {
	defer func() { err = translateError(err, attendance.ErrCorrection) }()
	return s.proposeCorrection(ctx, user, "", c, now)
}

func (s *Attendance) ProposeCorrectionOnce(ctx context.Context, user, key string, c attendance.Correction, now time.Time) (_ string, err error) {
	defer func() { err = translateError(err, attendance.ErrCorrection) }()
	if !attendance.ValidCorrectionKey(key) {
		return "", attendance.ErrCorrection
	}
	return s.proposeCorrection(ctx, user, key, c, now)
}

type correctionReceipt struct {
	Version      int    `firestore:"version"`
	PayloadHash  string `firestore:"payloadHash"`
	CorrectionID string `firestore:"correctionID"`
}

// correctionPayloadHash excludes server-owned fields and normalizes equivalent
// timestamp offsets and empty break lists before encoding the versioned payload.
func correctionPayloadHash(c attendance.Correction) (string, error) {
	c.ID, c.UserKey, c.CreatedAt = "", "", time.Time{}
	c.Start, c.End = c.Start.UTC(), c.End.UTC()
	c.Breaks = append([]worktime.Interval{}, c.Breaks...)
	for i := range c.Breaks {
		c.Breaks[i].Start, c.Breaks[i].End = c.Breaks[i].Start.UTC(), c.Breaks[i].End.UTC()
	}
	body, err := json.Marshal(struct {
		Version int
		Payload correctionDocument
	}{1, correctionDocumentFrom(c)})
	if err != nil {
		return "", attendance.ErrCorrection
	}
	sum := sha256.Sum256(body)
	return hex.EncodeToString(sum[:]), nil
}

func (s *Attendance) proposeCorrection(ctx context.Context, user, key string, c attendance.Correction, now time.Time) (string, error) {
	hash, err := correctionPayloadHash(c)
	if err != nil {
		return "", err
	}
	c.UserKey, c.CreatedAt = documentKey(user), now
	proposals := s.Client.Collection("channels").Doc(s.ChannelID).Collection("attendanceCorrectionRequests")
	var ref *db.DocumentRef
	var receipt *db.DocumentRef
	if key != "" {
		// Both paths are scoped by the authenticated user and channel. Never accept
		// a caller-provided actor or proposal ID as the source of this scope.
		receipt = s.user(user).Collection("attendanceCorrectionReceipts").Doc(documentKey(key))
	}
	err = s.Client.RunTransaction(ctx, func(ctx context.Context, tx *db.Transaction) error {
		if receipt != nil {
			prior, err := tx.Get(receipt)
			if err == nil {
				var stored correctionReceipt
				if err := prior.DataTo(&stored); err != nil {
					return err
				}
				if stored.Version != 1 || stored.PayloadHash != hash {
					return attendance.ErrIdempotencyConflict
				}
				if stored.CorrectionID == "" {
					return attendance.ErrCorrection
				}
				c.ID = stored.CorrectionID
				// A committed receipt remains replayable after the record is revised or
				// approved. It acknowledges the submission, not its current approval status.
				return nil
			}
			if !isMissingCorrection(err) {
				return err
			}
		}
		if ref == nil {
			id, err := uuid.NewRandom()
			if err != nil {
				return err
			}
			ref = proposals.Doc(id.String())
		}
		c.ID = ref.ID
		snapshot, err := tx.Get(s.user(user).Collection("attendance").Doc(c.RecordID))
		if err != nil {
			return err
		}
		var before attendance.Record
		if err := decodeRecord(snapshot, &before); err != nil {
			return err
		}
		before.RecordID = c.RecordID
		if c.End.After(now) {
			return attendance.ErrCorrection
		}
		if _, err := attendance.ApplyCorrection(before, c); err != nil {
			return err
		}
		if err := tx.Create(ref, correctionDocumentFrom(c)); err != nil {
			return err
		}
		if receipt != nil {
			return tx.Create(receipt, correctionReceipt{Version: 1, PayloadHash: hash, CorrectionID: c.ID})
		}
		return nil
	})
	if err != nil {
		return "", err
	}
	return c.ID, nil
}

// ApproveCorrection requires a trusted, pre-provisioned group reviewer and
// writes an immutable before/after audit document alongside the new projection.
func (s *Attendance) ApproveCorrection(ctx context.Context, reviewer, id string, now time.Time) (err error) {
	defer func() { err = translateError(err, attendance.ErrCorrection) }()
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
		if err := decodeCorrection(proposal, &c); err != nil {
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
		if err := decodeRecord(snapshot, &before); err != nil {
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
		if err := tx.Set(ref, recordDocumentFrom(after)); err != nil {
			return err
		}
		return tx.Create(audit, map[string]any{"proposal": correctionDocumentFrom(c), "before": recordDocumentFrom(before), "after": recordDocumentFrom(after), "reviewerKey": documentKey(reviewer), "approvedAt": now})
	})
}

func isMissingCorrection(err error) bool { return status.Code(err) == codes.NotFound }

func (s *Attendance) CorrectionStatus(ctx context.Context, user, id string) (_ string, err error) {
	defer func() { err = translateError(err, attendance.ErrCorrection) }()
	channel := s.Client.Collection("channels").Doc(s.ChannelID)
	result := "pending"
	err = s.Client.RunTransaction(ctx, func(ctx context.Context, tx *db.Transaction) error {
		result = "pending"
		snapshot, err := tx.Get(channel.Collection("attendanceCorrectionRequests").Doc(id))
		if err != nil {
			return err
		}
		var c attendance.Correction
		if err := decodeCorrection(snapshot, &c); err != nil {
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
		if err := decodeRecord(record, &current); err != nil {
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
	var stored struct {
		Proposal    correctionDocument `firestore:"proposal"`
		Before      recordDocument     `firestore:"before"`
		After       recordDocument     `firestore:"after"`
		ReviewerKey string             `firestore:"reviewerKey"`
		ApprovedAt  time.Time          `firestore:"approvedAt"`
	}
	if err := snapshot.DataTo(&stored); err != nil {
		return audit, err
	}
	audit = correctionAudit{Proposal: stored.Proposal.domain(), Before: stored.Before.domain(), After: stored.After.domain(), ReviewerKey: stored.ReviewerKey, ApprovedAt: stored.ApprovedAt}
	if audit.Proposal.UserKey != documentKey(user) || audit.Proposal.ID != id || audit.ReviewerKey == "" || audit.ReviewerKey == audit.Proposal.UserKey || audit.ApprovedAt.IsZero() {
		return audit, attendance.ErrCorrection
	}
	return audit, nil
}

// CorrectionReview exposes proposal details only to a group-authorized reviewer.
func (s *Attendance) CorrectionReview(ctx context.Context, reviewer, id string) (_ attendance.Correction, err error) {
	defer func() { err = translateError(err, attendance.ErrCorrection) }()
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
	if err := decodeCorrection(snapshot, &c); err != nil {
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
