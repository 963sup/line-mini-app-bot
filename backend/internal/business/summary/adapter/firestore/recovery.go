package firestore

import (
	"context"
	"errors"
	"log/slog"
	"strings"
	"time"

	db "cloud.google.com/go/firestore"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// RecoverPending also discovers legacy jobs committed before task creation. Group and job
// document IDs are hex digests. Bound the collection-group scan to this channel.
func (s *Summary) RecoverPending(ctx context.Context, run func(context.Context, string, string) error) error {
	channel := s.Client.Collection("channels").Doc(s.ChannelID)
	groups := channel.Collection("summaryGroups")
	bound := func(id string) *db.DocumentRef { return groups.Doc(id).Collection("summaryJobs").Doc("!") }
	query := s.Client.CollectionGroup("summaryJobs").OrderBy(db.DocumentID, db.Asc).EndBefore(bound("~")).Limit(50)
	cursor := channel.Collection("deliveryCursors").Doc("summary")
	start := query.StartAt(bound("!"))
	saved, err := cursor.Get(ctx)
	if err != nil && status.Code(err) != codes.NotFound {
		return err
	}
	if err == nil {
		var pos struct{ Group, Job string }
		if err := saved.DataTo(&pos); err != nil {
			return err
		}
		start = query.StartAfter(groups.Doc(pos.Group).Collection("summaryJobs").Doc(pos.Job))
	}
	docs, err := start.Documents(ctx).GetAll()
	if err != nil {
		return err
	}
	if len(docs) == 0 && saved != nil {
		_, err = cursor.Delete(ctx)
		return err
	}
	var failures []error
	checkpoint := func(doc *db.DocumentSnapshot) error {
		_, err := cursor.Set(ctx, map[string]string{"Group": doc.Ref.Parent.Parent.ID, "Job": doc.Ref.ID})
		return err
	}
	for _, doc := range docs {
		if err := ctx.Err(); err != nil {
			return err
		}
		group := doc.Ref.Parent.Parent
		if group == nil || group.Parent.Path != groups.Path {
			return errors.New("summary recovery outside configured channel")
		}
		var job jobDocument
		if err := doc.DataTo(&job); err != nil || job.ID != doc.Ref.ID || job.GroupID == "" || strings.Contains(job.GroupID, "/") || s.group(job.GroupID).Path != group.Path {
			failures = append(failures, errors.New("invalid summary recovery document"))
			continue
		}
		if job.Status != "pending" || time.Since(job.Cutoff) < 2*time.Minute || job.LeaseUntil.After(time.Now()) {
			continue
		}
		if err := checkpoint(doc); err != nil {
			return err
		}
		if err := run(ctx, job.GroupID, job.ID); err != nil {
			slog.WarnContext(ctx, "summary recovery deferred", "job_id", doc.Ref.ID, "age_seconds", int(time.Since(job.Cutoff).Seconds()))
			failures = append(failures, err)
		}
	}
	if len(docs) > 0 {
		failures = append(failures, checkpoint(docs[len(docs)-1]))
	}
	return errors.Join(failures...)
}
