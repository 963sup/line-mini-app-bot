package firestore

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"slices"
	"strings"
	"time"

	db "cloud.google.com/go/firestore"
	"github.com/963sup/line-mini-app-bot/internal/business/attendance/approvals"
	correctionsmodel "github.com/963sup/line-mini-app-bot/internal/business/attendance/corrections"
	recordmodel "github.com/963sup/line-mini-app-bot/internal/business/attendance/records"
	"github.com/963sup/line-mini-app-bot/internal/business/attendance/worktime"
	"github.com/google/uuid"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// ProposeCorrection preserves the existing internal caller contract. New Web
// submissions use ProposeCorrectionOnce so a lost response can be retried safely.
func (s *Attendance) ProposeCorrection(ctx context.Context, user string, c correctionsmodel.Correction, now time.Time) (_ string, err error) {
	defer func() { err = translateError(err, correctionsmodel.ErrCorrection) }()
	return s.proposeCorrection(ctx, user, "", c, now)
}

func (s *Attendance) ProposeCorrectionOnce(ctx context.Context, user, key string, c correctionsmodel.Correction, now time.Time) (_ string, err error) {
	defer func() { err = translateError(err, correctionsmodel.ErrCorrection) }()
	if !correctionsmodel.ValidCorrectionKey(key) {
		return "", correctionsmodel.ErrCorrection
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
func correctionPayloadHash(c correctionsmodel.Correction) (string, error) {
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
		return "", correctionsmodel.ErrCorrection
	}
	sum := sha256.Sum256(body)
	return hex.EncodeToString(sum[:]), nil
}

func (s *Attendance) proposeCorrection(ctx context.Context, user, key string, c correctionsmodel.Correction, now time.Time) (string, error) {
	hash, err := correctionPayloadHash(c)
	if err != nil {
		return "", err
	}
	if user == "" || !validSegment(c.GroupID) || len(c.GroupID) > 128 {
		return "", correctionsmodel.ErrCorrection
	}
	c.NewRecord = c.RecordID == ""
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
					return correctionsmodel.ErrIdempotencyConflict
				}
				if stored.CorrectionID == "" {
					return correctionsmodel.ErrCorrection
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
		if c.NewRecord {
			c.RecordID = "correction-" + c.ID
		}
		before, err := correctionBefore(tx, s.user(user), c)
		if err != nil {
			return err
		}

		if c.End.After(now) {
			return correctionsmodel.ErrCorrection
		}
		if _, err := correctionsmodel.ApplyCorrection(before, c); err != nil {
			return err
		}
		if err := correctionOverlap(tx, s.user(user), c); err != nil {
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

// ApproveCorrection and RejectCorrection share one atomic, terminal decision.
func (s *Attendance) ApproveCorrection(ctx context.Context, reviewer, id string, now time.Time) error {
	return s.decideCorrection(ctx, reviewer, id, "approved", "", now)
}

func (s *Attendance) RejectCorrection(ctx context.Context, reviewer, id, reason string, now time.Time) error {
	reason = strings.TrimSpace(reason)
	if reason == "" || len([]rune(reason)) > 500 {
		return correctionsmodel.ErrCorrection
	}
	return s.decideCorrection(ctx, reviewer, id, "rejected", reason, now)
}

func (s *Attendance) CancelCorrection(ctx context.Context, user, id string, now time.Time) error {
	return s.decideCorrection(ctx, user, id, "cancelled", "", now)
}

type correctionDecision struct {
	Status      string    `firestore:"status"`
	Reason      string    `firestore:"decisionReason"`
	ReviewerKey string    `firestore:"reviewerKey"`
	ReviewedAt  time.Time `firestore:"reviewedAt"`
}

func correctionBefore(tx *db.Transaction, employee *db.DocumentRef, c correctionsmodel.Correction) (recordmodel.Record, error) {
	snapshot, err := tx.Get(employee.Collection("attendance").Doc(c.RecordID))
	if c.NewRecord {
		if isMissingCorrection(err) {
			return recordmodel.Record{}, nil
		}
		if err == nil {
			return recordmodel.Record{}, correctionsmodel.ErrCorrection
		}
		return recordmodel.Record{}, err
	}
	if err != nil {
		return recordmodel.Record{}, err
	}
	var before recordmodel.Record
	if err := decodeRecord(snapshot, &before); err != nil {
		return before, err
	}
	before.RecordID = c.RecordID
	return before, nil
}

// Bound the equality-only scan and fail closed. The shared group state document
// serializes approvals with each other and with live punches, including inserts.
func correctionOverlap(tx *db.Transaction, employee *db.DocumentRef, c correctionsmodel.Correction) error {
	docs, err := tx.Documents(employee.Collection("attendance").Where("groupID", "==", c.GroupID).Limit(1001)).GetAll()
	if err != nil {
		return err
	}
	if len(docs) > 1000 {
		return correctionsmodel.ErrCorrection
	}
	for _, doc := range docs {
		if doc.Ref.ID == c.RecordID {
			continue
		}
		var other recordmodel.Record
		if err := decodeRecord(doc, &other); err != nil {
			return err
		}
		if other.ClockInAt.IsZero() || (other.ClockInAt.Before(c.End) && (other.ClockOutAt.IsZero() || other.ClockOutAt.After(c.Start))) {
			return correctionsmodel.ErrCorrection
		}
	}
	return nil
}

func (s *Attendance) reviewerGroups(tx *db.Transaction, reviewer string) ([]string, error) {
	channel := s.Client.Collection("channels").Doc(s.ChannelID)
	role, err := tx.Get(channel.Collection("attendanceReviewers").Doc(documentKey(reviewer)))
	if isMissingCorrection(err) {
		return nil, correctionsmodel.ErrCorrection
	}
	if err != nil {
		return nil, err
	}
	var access struct {
		Groups []string `firestore:"groups"`
	}
	if err := role.DataTo(&access); err != nil {
		return nil, err
	}
	return access.Groups, nil
}

func (s *Attendance) decideCorrection(ctx context.Context, actor, id, decision, reason string, now time.Time) (err error) {
	defer func() { err = translateError(err, correctionsmodel.ErrCorrection) }()
	if actor == "" || !validSegment(id) {
		return correctionsmodel.ErrCorrection
	}
	channel := s.Client.Collection("channels").Doc(s.ChannelID)
	return s.Client.RunTransaction(ctx, func(_ context.Context, tx *db.Transaction) error {
		proposalRef := channel.Collection("attendanceCorrectionRequests").Doc(id)
		proposal, err := tx.Get(proposalRef)
		if err != nil {
			return err
		}
		var c correctionsmodel.Correction
		if err := decodeCorrection(proposal, &c); err != nil {
			return err
		}
		if decision == "cancelled" {
			if c.UserKey != documentKey(actor) {
				return correctionsmodel.ErrCorrection
			}
		} else {
			groups, err := s.reviewerGroups(tx, actor)
			if err != nil {
				return err
			}
			if !approvals.CanReview(groups, documentKey(actor), c.UserKey, c.GroupID) {
				return correctionsmodel.ErrCorrection
			}
		}
		var prior correctionDecision
		if err := proposal.DataTo(&prior); err != nil {
			return err
		}
		audit := channel.Collection("attendanceCorrections").Doc(id)
		approved, err := tx.Get(audit)
		if err == nil && approved.Exists() {
			if decision == "approved" {
				return nil
			}
			return correctionsmodel.ErrCorrection
		}
		if err != nil && !isMissingCorrection(err) {
			return err
		}
		if prior.Status != "" {
			if prior.Status == decision && prior.Reason == reason {
				return nil
			}
			return correctionsmodel.ErrCorrection
		}
		if decision == "approved" {
			employee := channel.Collection("users").Doc(c.UserKey)
			before, err := correctionBefore(tx, employee, c)
			if err != nil {
				return err
			}
			if c.End.After(now) {
				return correctionsmodel.ErrCorrection
			}
			after, err := correctionsmodel.ApplyCorrection(before, c)
			if err != nil {
				return err
			}
			state := employee.Collection("attendanceStates").Doc(documentKey(c.GroupID))
			stateSnapshot, err := tx.Get(state)
			var active struct {
				RecordID string `firestore:"recordID"`
			}
			if err == nil {
				if err := stateSnapshot.DataTo(&active); err != nil {
					return err
				}
			} else if !isMissingCorrection(err) {
				return err
			}
			if err := correctionOverlap(tx, employee, c); err != nil {
				return err
			}
			update := map[string]any{"correctionUpdatedAt": now}
			if active.RecordID == c.RecordID {
				update["recordID"] = ""
			}
			if err := tx.Set(state, update, db.MergeAll); err != nil {
				return err
			}
			ref := employee.Collection("attendance").Doc(c.RecordID)
			if c.NewRecord {
				err = tx.Create(ref, recordDocumentFrom(after))
			} else {
				err = tx.Set(ref, recordDocumentFrom(after))
			}
			if err != nil {
				return err
			}
			if err := tx.Set(employee.Collection("attendanceGroups").Doc(documentKey(c.GroupID)), map[string]any{"groupID": c.GroupID}); err != nil {
				return err
			}
			if err := tx.Create(audit, map[string]any{"proposal": correctionDocumentFrom(c), "before": recordDocumentFrom(before), "after": recordDocumentFrom(after), "reviewerKey": documentKey(actor), "approvedAt": now}); err != nil {
				return err
			}
		}
		return tx.Update(proposalRef, []db.Update{{Path: "status", Value: decision}, {Path: "decisionReason", Value: reason}, {Path: "reviewerKey", Value: documentKey(actor)}, {Path: "reviewedAt", Value: now}})
	})
}

func isMissingCorrection(err error) bool { return status.Code(err) == codes.NotFound }

func (s *Attendance) CorrectionStatus(ctx context.Context, user, id string) (string, error) {
	details, err := s.CorrectionDetails(ctx, user, id, false)
	return details.Status, err
}

func (s *Attendance) correctionDetails(tx *db.Transaction, proposal *db.DocumentSnapshot, c correctionsmodel.Correction) (correctionsmodel.CorrectionDetails, error) {
	result := correctionsmodel.CorrectionDetails{Correction: c, Status: "pending"}
	var decision correctionDecision
	if err := proposal.DataTo(&decision); err != nil {
		return result, err
	}
	result.DecisionReason, result.ReviewedAt = decision.Reason, decision.ReviewedAt
	channel := s.Client.Collection("channels").Doc(s.ChannelID)
	audit, err := tx.Get(channel.Collection("attendanceCorrections").Doc(c.ID))
	if err == nil {
		evidence, err := decodeCorrectionAudit(audit, c.UserKey, c.ID)
		if err != nil {
			if err == correctionsmodel.ErrCorrection {
				result.Status = "invalid"
				return result, nil
			}
			return result, err
		}
		result.Before, result.ReviewedAt, result.Status = &evidence.Before, evidence.ApprovedAt, "approved"
		if !approvals.VerifyCorrection(evidence.Before, evidence.Proposal, evidence.After, evidence.After) {
			result.Status = "invalid"
		}
		return result, nil
	}
	if !isMissingCorrection(err) {
		return result, err
	}
	if decision.Status == "rejected" || decision.Status == "cancelled" {
		result.Status = decision.Status
		return result, nil
	}
	if decision.Status != "" {
		result.Status = "invalid"
		return result, nil
	}
	employee := channel.Collection("users").Doc(c.UserKey)
	before, err := correctionBefore(tx, employee, c)
	if err != nil {
		if err == correctionsmodel.ErrCorrection || isMissingCorrection(err) {
			result.Status = "stale"
			return result, nil
		}
		return result, err
	}
	if !c.NewRecord {
		result.Before = &before
	}
	if _, err := correctionsmodel.ApplyCorrection(before, c); err != nil {
		result.Status = "stale"
	}
	return result, nil
}

func (s *Attendance) CorrectionDetails(ctx context.Context, actor, id string, review bool) (_ correctionsmodel.CorrectionDetails, err error) {
	defer func() { err = translateError(err, correctionsmodel.ErrCorrection) }()
	var result correctionsmodel.CorrectionDetails
	err = s.Client.RunTransaction(ctx, func(_ context.Context, tx *db.Transaction) error {
		proposal, err := tx.Get(s.Client.Collection("channels").Doc(s.ChannelID).Collection("attendanceCorrectionRequests").Doc(id))
		if err != nil {
			return err
		}
		var c correctionsmodel.Correction
		if err := decodeCorrection(proposal, &c); err != nil {
			return err
		}
		if review {
			groups, err := s.reviewerGroups(tx, actor)
			if err != nil {
				return err
			}
			if !approvals.CanReview(groups, documentKey(actor), c.UserKey, c.GroupID) {
				return correctionsmodel.ErrCorrection
			}
		} else if c.UserKey != documentKey(actor) {
			return correctionsmodel.ErrCorrection
		}
		result, err = s.correctionDetails(tx, proposal, c)
		return err
	}, db.ReadOnly)
	return result, err
}

func (s *Attendance) CorrectionsPage(ctx context.Context, actor, group, cursor string, review bool, size int) (_ correctionsmodel.CorrectionPage, err error) {
	defer func() { err = translateError(err, correctionsmodel.ErrCorrection) }()
	result := correctionsmodel.CorrectionPage{Items: []correctionsmodel.CorrectionDetails{}}
	if !validPageSize(size) || (review && !validSegment(group)) {
		return result, recordmodel.ErrInvalidCursor
	}
	scope := s.cursorScope("corrections-own", actor, "", time.Time{})
	if review {
		scope = s.cursorScope("corrections-review", actor, group, time.Time{})
	}
	c, err := decodeCursor(cursor, scope)
	if err != nil {
		return result, err
	}
	err = s.Client.RunTransaction(ctx, func(_ context.Context, tx *db.Transaction) error {
		result = correctionsmodel.CorrectionPage{Items: []correctionsmodel.CorrectionDetails{}}
		q := s.Client.Collection("channels").Doc(s.ChannelID).Collection("attendanceCorrectionRequests").Where("userKey", "==", documentKey(actor))
		if review {
			groups, err := s.reviewerGroups(tx, actor)
			if err != nil {
				return err
			}
			if !slices.Contains(groups, group) {
				return correctionsmodel.ErrCorrection
			}
			q = s.Client.Collection("channels").Doc(s.ChannelID).Collection("attendanceCorrectionRequests").Where("groupID", "==", group)
		}
		q = q.OrderBy(db.DocumentID, db.Desc)
		if c.ID != "" {
			q = q.StartAfter(c.ID)
		}
		docs, err := tx.Documents(q.Limit(size + 1)).GetAll()
		if err != nil {
			return err
		}
		if len(docs) > size {
			result.NextCursor = encodeCursor(scope, time.Time{}, docs[size-1].Ref.ID)
			docs = docs[:size]
		}
		for _, doc := range docs {
			var proposal correctionsmodel.Correction
			if err := decodeCorrection(doc, &proposal); err != nil {
				return err
			}
			if review && proposal.UserKey == documentKey(actor) {
				continue
			}
			details, err := s.correctionDetails(tx, doc, proposal)
			if err != nil {
				return err
			}
			result.Items = append(result.Items, details)
		}
		return nil
	}, db.ReadOnly)
	return result, err
}

func (s *Attendance) verifyCorrection(ctx context.Context, user string, current recordmodel.Record) (bool, error) {
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
	audit, err := decodeCorrectionAudit(snapshot, documentKey(user), current.CorrectionID)
	if err != nil {
		if err == correctionsmodel.ErrCorrection {
			return false, nil
		}
		return false, err
	}
	return approvals.VerifyCorrection(audit.Before, audit.Proposal, audit.After, current), nil
}

type correctionAudit struct {
	Proposal    correctionsmodel.Correction `firestore:"proposal"`
	Before      recordmodel.Record          `firestore:"before"`
	After       recordmodel.Record          `firestore:"after"`
	ReviewerKey string                      `firestore:"reviewerKey"`
	ApprovedAt  time.Time                   `firestore:"approvedAt"`
}

func decodeCorrectionAudit(snapshot *db.DocumentSnapshot, userKey, id string) (correctionAudit, error) {
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
	if audit.Proposal.UserKey != userKey || audit.Proposal.ID != id || audit.ReviewerKey == "" || audit.ReviewerKey == audit.Proposal.UserKey || audit.ApprovedAt.IsZero() {
		return audit, correctionsmodel.ErrCorrection
	}
	return audit, nil
}

func (s *Attendance) ReviewerGroups(ctx context.Context, actor string) (_ []string, err error) {
	defer func() { err = translateError(err, correctionsmodel.ErrCorrection) }()
	var groups []string
	err = s.Client.RunTransaction(ctx, func(_ context.Context, tx *db.Transaction) error {
		var err error
		groups, err = s.reviewerGroups(tx, actor)
		return err
	}, db.ReadOnly)
	return groups, err
}
