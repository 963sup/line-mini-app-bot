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
	attendancecommand "github.com/963sup/line-mini-app-bot/internal/business/attendance/application/command"
	attendancedto "github.com/963sup/line-mini-app-bot/internal/business/attendance/application/dto"
	attendanceentity "github.com/963sup/line-mini-app-bot/internal/business/attendance/domain/entity"
	attendancepolicy "github.com/963sup/line-mini-app-bot/internal/business/attendance/domain/policy"
	attendanceservice "github.com/963sup/line-mini-app-bot/internal/business/attendance/domain/service"
	attendancevalueobject "github.com/963sup/line-mini-app-bot/internal/business/attendance/domain/valueobject"
	"github.com/963sup/line-mini-app-bot/internal/platform/firestore/key"
	"github.com/google/uuid"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// ProposeCorrection preserves the existing internal caller contract. New Web
// submissions use ProposeCorrectionOnce so a lost response can be retried safely.
func (s *Attendance) ProposeCorrection(ctx context.Context, user string, c attendanceentity.Correction, now time.Time) (_ string, err error) {
	defer func() { err = translateError(err, attendanceentity.ErrCorrection) }()
	return s.proposeCorrection(ctx, user, "", c, now)
}

func (s *Attendance) ProposeCorrectionOnce(ctx context.Context, user, key string, c attendanceentity.Correction, now time.Time) (_ string, err error) {
	defer func() { err = translateError(err, attendanceentity.ErrCorrection) }()
	if !attendancevalueobject.ValidCorrectionKey(key) {
		return "", attendanceentity.ErrCorrection
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
func correctionPayloadHash(c attendanceentity.Correction) (string, error) {
	c.ID, c.UserKey, c.CreatedAt = "", "", time.Time{}
	c.Start, c.End = c.Start.UTC(), c.End.UTC()
	c.Breaks = append([]attendancevalueobject.Interval{}, c.Breaks...)
	for i := range c.Breaks {
		c.Breaks[i].Start, c.Breaks[i].End = c.Breaks[i].Start.UTC(), c.Breaks[i].End.UTC()
	}
	body, err := json.Marshal(struct {
		Version int
		Payload correctionDocument
	}{1, correctionDocumentFrom(c)})
	if err != nil {
		return "", attendanceentity.ErrCorrection
	}
	sum := sha256.Sum256(body)
	return hex.EncodeToString(sum[:]), nil
}

func (s *Attendance) proposeCorrection(ctx context.Context, user, key string, c attendanceentity.Correction, now time.Time) (string, error) {
	hash, err := correctionPayloadHash(c)
	if err != nil {
		return "", err
	}
	if user == "" || !validSegment(c.GroupID) || len(c.GroupID) > 128 {
		return "", attendanceentity.ErrCorrection
	}
	c.NewRecord = c.RecordID == ""
	c.UserKey, c.CreatedAt = firestorekey.Encode(user), now
	proposals := s.Client.Collection("channels").Doc(s.ChannelID).Collection("attendanceCorrectionRequests")
	var ref *db.DocumentRef
	var receipt *db.DocumentRef
	if key != "" {
		// Both paths are scoped by the authenticated user and channel. Never accept
		// a caller-provided actor or proposal ID as the source of this scope.
		receipt = s.user(user).Collection("attendanceCorrectionReceipts").Doc(firestorekey.Encode(key))
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
					return attendanceentity.ErrIdempotencyConflict
				}
				if stored.CorrectionID == "" {
					return attendanceentity.ErrCorrection
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
			return attendanceentity.ErrCorrection
		}
		if _, err := attendanceservice.ApplyCorrection(before, c); err != nil {
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
		return attendanceentity.ErrCorrection
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

func correctionBefore(tx *db.Transaction, employee *db.DocumentRef, c attendanceentity.Correction) (attendanceentity.Record, error) {
	snapshot, err := tx.Get(employee.Collection("attendance").Doc(c.RecordID))
	if c.NewRecord {
		if isMissingCorrection(err) {
			return attendanceentity.Record{}, nil
		}
		if err == nil {
			return attendanceentity.Record{}, attendanceentity.ErrCorrection
		}
		return attendanceentity.Record{}, err
	}
	if err != nil {
		return attendanceentity.Record{}, err
	}
	var before attendanceentity.Record
	if err := decodeRecord(snapshot, &before); err != nil {
		return before, err
	}
	before.RecordID = c.RecordID
	return before, nil
}

// Bound the equality-only scan and fail closed. The shared group state document
// serializes approvals with each other and with live punches, including inserts.
func correctionOverlap(tx *db.Transaction, employee *db.DocumentRef, c attendanceentity.Correction) error {
	docs, err := tx.Documents(employee.Collection("attendance").Where("groupID", "==", c.GroupID).Limit(1001)).GetAll()
	if err != nil {
		return err
	}
	if len(docs) > 1000 {
		return attendanceentity.ErrCorrection
	}
	for _, doc := range docs {
		if doc.Ref.ID == c.RecordID {
			continue
		}
		var other attendanceentity.Record
		if err := decodeRecord(doc, &other); err != nil {
			return err
		}
		if other.ClockInAt.IsZero() || (other.ClockInAt.Before(c.End) && (other.ClockOutAt.IsZero() || other.ClockOutAt.After(c.Start))) {
			return attendanceentity.ErrCorrection
		}
	}
	return nil
}

func (s *Attendance) reviewerGroups(tx *db.Transaction, reviewer string) ([]string, error) {
	channel := s.Client.Collection("channels").Doc(s.ChannelID)
	role, err := tx.Get(channel.Collection("attendanceReviewers").Doc(firestorekey.Encode(reviewer)))
	if isMissingCorrection(err) {
		return nil, attendanceentity.ErrCorrection
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
	defer func() { err = translateError(err, attendanceentity.ErrCorrection) }()
	if actor == "" || !validSegment(id) {
		return attendanceentity.ErrCorrection
	}
	channel := s.Client.Collection("channels").Doc(s.ChannelID)
	return s.Client.RunTransaction(ctx, func(_ context.Context, tx *db.Transaction) error {
		proposalRef := channel.Collection("attendanceCorrectionRequests").Doc(id)
		proposal, err := tx.Get(proposalRef)
		if err != nil {
			return err
		}
		var c attendanceentity.Correction
		if err := decodeCorrection(proposal, &c); err != nil {
			return err
		}
		t := reviewTransaction{store: s, tx: tx, proposal: proposal, audit: channel.Collection("attendanceCorrections").Doc(id), employee: channel.Collection("users").Doc(c.UserKey), actor: actor}
		return attendancecommand.ReviewCorrection(t, firestorekey.Encode(actor), c, attendancepolicy.Decision{Status: decision, Reason: reason}, now)
	})
}

type reviewTransaction struct {
	store           *Attendance
	tx              *db.Transaction
	proposal        *db.DocumentSnapshot
	audit, employee *db.DocumentRef
	actor           string
}

func (t reviewTransaction) ReviewerGroups() ([]string, error) {
	return t.store.reviewerGroups(t.tx, t.actor)
}

func (t reviewTransaction) PreviousDecision() (attendancepolicy.Decision, bool, error) {
	var prior correctionDecision
	if err := t.proposal.DataTo(&prior); err != nil {
		return attendancepolicy.Decision{}, false, err
	}
	approved, err := t.tx.Get(t.audit)
	if err != nil && !isMissingCorrection(err) {
		return attendancepolicy.Decision{}, false, err
	}
	return attendancepolicy.Decision{Status: prior.Status, Reason: prior.Reason}, err == nil && approved.Exists(), nil
}

func (t reviewTransaction) Original(c attendanceentity.Correction) (attendanceentity.Record, error) {
	return correctionBefore(t.tx, t.employee, c)
}

func (t reviewTransaction) SaveApproval(c attendanceentity.Correction, before, after attendanceentity.Record, actorKey string, now time.Time) error {
	employee := t.employee
	state := employee.Collection("attendanceStates").Doc(firestorekey.Encode(c.GroupID))
	stateSnapshot, err := t.tx.Get(state)
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
	if err := correctionOverlap(t.tx, employee, c); err != nil {
		return err
	}
	update := map[string]any{"correctionUpdatedAt": now}
	if active.RecordID == c.RecordID {
		update["recordID"] = ""
	}
	if err := t.tx.Set(state, update, db.MergeAll); err != nil {
		return err
	}
	ref := employee.Collection("attendance").Doc(c.RecordID)
	if c.NewRecord {
		err = t.tx.Create(ref, recordDocumentFrom(after))
	} else {
		err = t.tx.Set(ref, recordDocumentFrom(after))
	}
	if err != nil {
		return err
	}
	if err := t.tx.Set(employee.Collection("attendanceGroups").Doc(firestorekey.Encode(c.GroupID)), map[string]any{"groupID": c.GroupID}); err != nil {
		return err
	}
	if err := t.tx.Create(t.audit, map[string]any{"proposal": correctionDocumentFrom(c), "before": recordDocumentFrom(before), "after": recordDocumentFrom(after), "reviewerKey": actorKey, "approvedAt": now}); err != nil {
		return err
	}
	return nil
}

func (t reviewTransaction) SaveDecision(d attendancepolicy.Decision, actorKey string, now time.Time) error {
	return t.tx.Update(t.proposal.Ref, []db.Update{{Path: "status", Value: d.Status}, {Path: "decisionReason", Value: d.Reason}, {Path: "reviewerKey", Value: actorKey}, {Path: "reviewedAt", Value: now}})
}
func isMissingCorrection(err error) bool { return status.Code(err) == codes.NotFound }

func (s *Attendance) CorrectionStatus(ctx context.Context, user, id string) (string, error) {
	details, err := s.CorrectionDetails(ctx, user, id, false)
	return details.Status, err
}

func (s *Attendance) correctionDetails(tx *db.Transaction, proposal *db.DocumentSnapshot, c attendanceentity.Correction) (attendancedto.CorrectionDetails, error) {
	result := attendancedto.CorrectionDetails{Correction: c, Status: "pending"}
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
			if err == attendanceentity.ErrCorrection {
				result.Status = "invalid"
				return result, nil
			}
			return result, err
		}
		result.Before, result.ReviewedAt, result.Status = &evidence.Before, evidence.ApprovedAt, "approved"
		if !attendanceservice.VerifyCorrection(evidence.Before, evidence.Proposal, evidence.After, evidence.After) {
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
		if err == attendanceentity.ErrCorrection || isMissingCorrection(err) {
			result.Status = "stale"
			return result, nil
		}
		return result, err
	}
	if !c.NewRecord {
		result.Before = &before
	}
	if _, err := attendanceservice.ApplyCorrection(before, c); err != nil {
		result.Status = "stale"
	}
	return result, nil
}

func (s *Attendance) CorrectionDetails(ctx context.Context, actor, id string, review bool) (_ attendancedto.CorrectionDetails, err error) {
	defer func() { err = translateError(err, attendanceentity.ErrCorrection) }()
	var result attendancedto.CorrectionDetails
	err = s.Client.RunTransaction(ctx, func(_ context.Context, tx *db.Transaction) error {
		proposal, err := tx.Get(s.Client.Collection("channels").Doc(s.ChannelID).Collection("attendanceCorrectionRequests").Doc(id))
		if err != nil {
			return err
		}
		var c attendanceentity.Correction
		if err := decodeCorrection(proposal, &c); err != nil {
			return err
		}
		if review {
			groups, err := s.reviewerGroups(tx, actor)
			if err != nil {
				return err
			}
			if !attendancepolicy.CanReview(groups, firestorekey.Encode(actor), c.UserKey, c.GroupID) {
				return attendanceentity.ErrCorrection
			}
		} else if c.UserKey != firestorekey.Encode(actor) {
			return attendanceentity.ErrCorrection
		}
		result, err = s.correctionDetails(tx, proposal, c)
		return err
	}, db.ReadOnly)
	return result, err
}

func (s *Attendance) CorrectionsPage(ctx context.Context, actor, group, cursor string, review bool, size int) (_ attendancedto.CorrectionPage, err error) {
	defer func() { err = translateError(err, attendanceentity.ErrCorrection) }()
	result := attendancedto.CorrectionPage{Items: []attendancedto.CorrectionDetails{}}
	if !validPageSize(size) || (review && !validSegment(group)) {
		return result, attendancedto.ErrInvalidCursor
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
		result = attendancedto.CorrectionPage{Items: []attendancedto.CorrectionDetails{}}
		q := s.Client.Collection("channels").Doc(s.ChannelID).Collection("attendanceCorrectionRequests").Where("userKey", "==", firestorekey.Encode(actor))
		if review {
			groups, err := s.reviewerGroups(tx, actor)
			if err != nil {
				return err
			}
			if !slices.Contains(groups, group) {
				return attendanceentity.ErrCorrection
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
			var proposal attendanceentity.Correction
			if err := decodeCorrection(doc, &proposal); err != nil {
				return err
			}
			if review && proposal.UserKey == firestorekey.Encode(actor) {
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

func (s *Attendance) verifyCorrection(ctx context.Context, user string, current attendanceentity.Record) (bool, error) {
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
	audit, err := decodeCorrectionAudit(snapshot, firestorekey.Encode(user), current.CorrectionID)
	if err != nil {
		if err == attendanceentity.ErrCorrection {
			return false, nil
		}
		return false, err
	}
	return attendanceservice.VerifyCorrection(audit.Before, audit.Proposal, audit.After, current), nil
}

type correctionAudit struct {
	Proposal    attendanceentity.Correction `firestore:"proposal"`
	Before      attendanceentity.Record     `firestore:"before"`
	After       attendanceentity.Record     `firestore:"after"`
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
		return audit, attendanceentity.ErrCorrection
	}
	return audit, nil
}

func (s *Attendance) ReviewerGroups(ctx context.Context, actor string) (_ []string, err error) {
	defer func() { err = translateError(err, attendanceentity.ErrCorrection) }()
	var groups []string
	err = s.Client.RunTransaction(ctx, func(_ context.Context, tx *db.Transaction) error {
		var err error
		groups, err = s.reviewerGroups(tx, actor)
		return err
	}, db.ReadOnly)
	return groups, err
}
