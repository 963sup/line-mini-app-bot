package firestore

import (
	"context"
	"errors"
	"fmt"
	"time"

	db "cloud.google.com/go/firestore"
	"github.com/963sup/line-mini-app-bot/internal/summary"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type Chat struct {
	Client    *db.Client
	ChannelID string
}
type chatState struct {
	Disabled                     bool
	Epoch, Revision, LastControl int64
}

// Edit only updates a retained original, preserving its original expiry and
// rejecting reordered edits. Revision invalidates in-flight summary results.
func (s *Chat) Edit(ctx context.Context, g, id, text string, timestamp int64) error {
	if id == "" || timestamp <= 0 || timestamp > time.Now().Add(time.Minute).UnixMilli() {
		return nil
	}
	root := s.group(g)
	ref := root.Collection("summaryMessages").Doc(key(id))
	return s.Client.RunTransaction(ctx, func(ctx context.Context, tx *db.Transaction) error {
		st, err := state(tx, root)
		if err != nil || st.Disabled {
			return err
		}
		doc, err := tx.Get(ref)
		if status.Code(err) == codes.NotFound {
			return nil
		}
		if err != nil {
			return err
		}
		var m summary.Message
		if err := doc.DataTo(&m); err != nil {
			return err
		}
		if m.Deleted || m.Epoch != st.Epoch || !m.ExpireAt.After(time.Now()) || timestamp <= m.EditedAt || timestamp < m.OccurredAt.UnixMilli() {
			return nil
		}
		m.Text, m.EditedAt = text, timestamp
		st.Revision++
		if err := tx.Set(root, st); err != nil {
			return err
		}
		return tx.Set(ref, m)
	})
}

func (s *Chat) group(g string) *db.DocumentRef {
	return s.Client.Collection("channels").Doc(s.ChannelID).Collection("summaryGroups").Doc(key(g))
}
func state(tx *db.Transaction, r *db.DocumentRef) (chatState, error) {
	var st chatState
	d, e := tx.Get(r)
	if status.Code(e) == codes.NotFound {
		return st, nil
	}
	if e != nil {
		return st, e
	}
	return st, d.DataTo(&st)
}
func (s *Chat) Collect(ctx context.Context, g string, m summary.Message) error {
	now := time.Now()
	if m.ID == "" || m.UserID == "" || m.OccurredAt.Before(now.Add(-24*time.Hour)) || m.OccurredAt.After(now.Add(time.Minute)) {
		return nil
	}
	root := s.group(g)
	ref := root.Collection("summaryMessages").Doc(key(m.ID))
	return s.Client.RunTransaction(ctx, func(ctx context.Context, tx *db.Transaction) error {
		st, e := state(tx, root)
		if e != nil {
			return e
		}
		if st.Disabled {
			return nil
		}
		_, e = tx.Get(ref)
		if e == nil {
			return nil
		}
		if status.Code(e) != codes.NotFound {
			return e
		}
		if m.OccurredAt.UnixMilli() < st.LastControl {
			return nil
		}
		m.Epoch = st.Epoch
		m.ExpireAt = m.OccurredAt.Add(24 * time.Hour)
		return tx.Create(ref, m)
	})
}
func (s *Chat) Unsend(ctx context.Context, g, id string) error {
	if id == "" {
		return nil
	}
	root := s.group(g)
	err := s.Client.RunTransaction(ctx, func(ctx context.Context, tx *db.Transaction) error {
		st, e := state(tx, root)
		if e != nil {
			return e
		}
		st.Revision++
		if e := tx.Set(root, st); e != nil {
			return e
		}
		// A tombstone prevents a delayed original message from restoring content.
		return tx.Set(root.Collection("summaryMessages").Doc(key(id)), summary.Message{ID: id, Deleted: true, Epoch: st.Epoch, ExpireAt: time.Now().Add(48 * time.Hour)})
	})
	if err != nil {
		return err
	}
	// Drop any cached result containing the withdrawn source. Generation checks
	// in mutate also prevent an in-flight model call from restoring the result.
	for {
		docs, e := root.Collection("summaryJobs").Where("Sources", "array-contains", id).Limit(100).Documents(ctx).GetAll()
		if e != nil {
			return e
		}
		if len(docs) == 0 {
			return nil
		}
		batch := s.Client.Batch()
		for _, d := range docs {
			batch.Update(d.Ref, []db.Update{{Path: "Result", Value: ""}, {Path: "Sources", Value: []string{}}, {Path: "Status", Value: "done"}})
		}
		if _, e = batch.Commit(ctx); e != nil {
			return e
		}
	}
}

func (s *Chat) Command(ctx context.Context, g, event, command string, timestamp int64) (summary.Job, error) {
	root := s.group(g)
	id := key(event)
	ref := root.Collection("summaryJobs").Doc(id)
	var job summary.Job
	now := time.Now().UTC()
	retry := summary.RandomID()
	err := s.Client.RunTransaction(ctx, func(ctx context.Context, tx *db.Transaction) error {
		old, e := tx.Get(ref)
		if e == nil {
			return old.DataTo(&job)
		}
		if status.Code(e) != codes.NotFound {
			return e
		}
		st, e := state(tx, root)
		if e != nil {
			return e
		}
		job = summary.Job{ID: id, GroupID: g, Kind: "summary", Status: "pending", Cutoff: now, ExpireAt: now.Add(24 * time.Hour), RetryKey: retry}
		if command != "摘要" {
			job.Kind = "purge"
			if timestamp < st.LastControl {
				job.Status = "done"
			} else {
				st.LastControl = timestamp
				st.Epoch++
				st.Revision++
				st.Disabled = command == "關閉摘要"
			}
		} else if st.Disabled {
			job.Status = "disabled"
		}
		job.Epoch, job.Revision = st.Epoch, st.Revision
		if e := tx.Set(root, st); e != nil {
			return e
		}
		return tx.Create(ref, job)
	})
	return job, err
}

func (s *Chat) Claim(ctx context.Context, g, id string) (summary.Job, bool, error) {
	var j summary.Job
	claimed := false
	root := s.group(g)
	ref := root.Collection("summaryJobs").Doc(id)
	now := time.Now()
	lease := summary.RandomID()
	err := s.Client.RunTransaction(ctx, func(ctx context.Context, tx *db.Transaction) error {
		claimed = false
		d, e := tx.Get(ref)
		if status.Code(e) == codes.NotFound {
			return nil
		}
		if e != nil {
			return e
		}
		if e = d.DataTo(&j); e != nil {
			return e
		}
		if j.Status == "done" || j.Status == "disabled" {
			return nil
		}
		if !j.ExpireAt.After(now) {
			return tx.Update(ref, []db.Update{{Path: "Status", Value: "done"}, {Path: "Result", Value: ""}})
		}
		if j.LeaseUntil.After(now) {
			return errors.New("summary job busy")
		}
		j.Lease = lease
		j.LeaseUntil = now.Add(2 * time.Minute)
		claimed = true
		return tx.Set(ref, j)
	})
	return j, claimed, err
}
func (s *Chat) Messages(ctx context.Context, j summary.Job) ([]summary.Message, error) {
	// One range/order field: no composite index required. Filter epochs and TTL in code.
	docs, e := s.group(j.GroupID).Collection("summaryMessages").Where("OccurredAt", ">=", j.Cutoff.Add(-24*time.Hour)).Where("OccurredAt", "<=", j.Cutoff).OrderBy("OccurredAt", db.Desc).Limit(100).Documents(ctx).GetAll()
	if e != nil {
		return nil, e
	}
	result := make([]summary.Message, 0, len(docs))
	now := time.Now()
	for i := len(docs) - 1; i >= 0; i-- {
		var m summary.Message
		if e := docs[i].DataTo(&m); e != nil {
			return nil, e
		}
		if m.Epoch == j.Epoch && !m.Deleted && m.ExpireAt.After(now) {
			result = append(result, m)
		}
	}
	return result, nil
}
func (s *Chat) mutate(ctx context.Context, j summary.Job, update func(*summary.Job)) error {
	ref := s.group(j.GroupID).Collection("summaryJobs").Doc(j.ID)
	return s.Client.RunTransaction(ctx, func(ctx context.Context, tx *db.Transaction) error {
		st, e := state(tx, s.group(j.GroupID))
		if e != nil {
			return e
		}
		d, e := tx.Get(ref)
		if e != nil {
			return e
		}
		var current summary.Job
		if e = d.DataTo(&current); e != nil {
			return e
		}
		if current.Lease != j.Lease {
			return fmt.Errorf("summary lease changed")
		}
		if current.Kind == "summary" && (st.Disabled || st.Epoch != current.Epoch || st.Revision != current.Revision) {
			current.Status = "done"
			current.Result = ""
			current.Sources = nil
		} else {
			update(&current)
		}
		return tx.Set(ref, current)
	})
}
func (s *Chat) SaveResult(ctx context.Context, j summary.Job, text string, sources []string) error {
	return s.mutate(ctx, j, func(v *summary.Job) { v.Result = text; v.Sources = sources; v.ExpireAt = j.ExpireAt })
}
func (s *Chat) Done(ctx context.Context, j summary.Job) error {
	return s.mutate(ctx, j, func(v *summary.Job) { v.Status = "done"; v.LeaseUntil = time.Time{}; v.Result = "" })
}
func (s *Chat) Current(ctx context.Context, j summary.Job) (bool, error) {
	d, e := s.group(j.GroupID).Get(ctx)
	if e != nil {
		return false, e
	}
	var st chatState
	if e = d.DataTo(&st); e != nil {
		return false, e
	}
	return !st.Disabled && st.Epoch == j.Epoch && st.Revision == j.Revision && j.ExpireAt.After(time.Now()), nil
}
func (s *Chat) Purge(ctx context.Context, j summary.Job) error {
	root := s.group(j.GroupID)
	for _, collection := range []string{"summaryMessages", "summaryJobs"} {
		for {
			docs, e := root.Collection(collection).Where("Epoch", "<", j.Epoch).Limit(200).Documents(ctx).GetAll()
			if e != nil {
				return e
			}
			if len(docs) == 0 {
				break
			}
			batch := s.Client.Batch()
			for _, d := range docs {
				batch.Delete(d.Ref)
			}
			if _, e = batch.Commit(ctx); e != nil {
				return e
			}
		}
	}
	return nil
}
