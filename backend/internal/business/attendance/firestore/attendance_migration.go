package firestore

import (
	"context"
	"errors"
	"time"

	db "cloud.google.com/go/firestore"
	recordmodel "github.com/963sup/line-mini-app-bot/internal/business/attendance/records"
	"google.golang.org/api/iterator"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// QueryMigration is a resumable checkpoint, scoped by the CLI to project/database/channel.
// A failed batch can be repeated: each update re-reads the current shift in a transaction.
type QueryMigration struct {
	Version      int      `json:"version"`
	ChannelID    string   `json:"channelId"`
	Apply        bool     `json:"apply"`
	UsersToken   string   `json:"usersToken,omitempty"`
	PendingUsers []string `json:"pendingUsers,omitempty"`
	AfterRecord  string   `json:"afterRecord,omitempty"`
	ListedUsers  bool     `json:"listedUsers"`
	Complete     bool     `json:"complete"`
	Records      int      `json:"records"`
}

// MigrateQueryBatch processes at most 200 shifts. Listing uses DocumentRefs to
// include missing user parents: historical writes only created their subcollections.
// All writers must maintain the new projection before the first apply run.
func (s *Attendance) MigrateQueryBatch(ctx context.Context, progress QueryMigration, apply bool) (QueryMigration, error) {
	if progress.Version == 0 {
		progress = QueryMigration{Version: 1, ChannelID: s.ChannelID, Apply: apply}
	}
	if progress.Version != 1 || progress.ChannelID != s.ChannelID || progress.Apply != apply {
		return progress, errors.New("migration checkpoint scope does not match")
	}
	if progress.Complete {
		return progress, nil
	}
	if len(progress.PendingUsers) == 0 && (!progress.ListedUsers || progress.UsersToken != "") {
		users := s.Client.Collection("channels").Doc(s.ChannelID).Collection("users").DocumentRefs(ctx)
		pager := iterator.NewPager(users, 20, progress.UsersToken)
		var refs []*db.DocumentRef
		token, err := pager.NextPage(&refs)
		if err != nil {
			return progress, err
		}
		progress.UsersToken = token
		progress.ListedUsers = true
		for _, ref := range refs {
			progress.PendingUsers = append(progress.PendingUsers, ref.ID)
		}
	}
	if len(progress.PendingUsers) == 0 {
		progress.Complete = true
		return progress, nil
	}
	userID := progress.PendingUsers[0]
	if !validSegment(userID) || (progress.AfterRecord != "" && !validSegment(progress.AfterRecord)) {
		return progress, errors.New("invalid migration checkpoint path")
	}
	user := s.Client.Collection("channels").Doc(s.ChannelID).Collection("users").Doc(userID)
	q := user.Collection("attendance").OrderBy(db.DocumentID, db.Asc)
	if progress.AfterRecord != "" {
		q = q.StartAfter(progress.AfterRecord)
	}
	docs, err := q.Limit(queryBatchSize).Documents(ctx).GetAll()
	if err != nil {
		return progress, err
	}
	for _, doc := range docs {
		if err := s.migrateQueryRecord(ctx, user, doc, apply); err != nil {
			return progress, err
		}
	}
	progress.Records += len(docs)
	if len(docs) < queryBatchSize {
		progress.PendingUsers = progress.PendingUsers[1:]
		progress.AfterRecord = ""
	} else {
		progress.AfterRecord = docs[len(docs)-1].Ref.ID
	}
	return progress, nil
}

func (s *Attendance) migrateQueryRecord(ctx context.Context, user *db.DocumentRef, snapshot *db.DocumentSnapshot, apply bool) error {
	inspect := func(doc *db.DocumentSnapshot) (recordmodel.Record, error) {
		var r recordmodel.Record
		if err := decodeRecord(doc, &r); err != nil {
			return r, err
		}
		if !validSegment(r.GroupID) || r.ClockInAt.IsZero() {
			return r, errors.New("attendance record has invalid group or start timestamp")
		}
		return r, nil
	}
	if !apply {
		_, err := inspect(snapshot)
		return err
	}
	return s.Client.RunTransaction(ctx, func(ctx context.Context, tx *db.Transaction) error {
		current, err := tx.Get(snapshot.Ref)
		if status.Code(err) == codes.NotFound {
			return nil
		}
		if err != nil {
			return err
		}
		r, err := inspect(current)
		if err != nil {
			return err
		}
		// Update only derived query metadata. Do not overwrite concurrent corrections or raw evidence.
		if err := tx.Update(current.Ref, []db.Update{{Path: "isOpen", Value: r.ClockOutAt.IsZero()}}); err != nil {
			return err
		}
		return tx.Set(user.Collection("attendanceGroups").Doc(documentKey(r.GroupID)), map[string]any{"groupID": r.GroupID})
	})
}

// ActivateQueries is a separate operator action after indexes and all dual-writing
// application revisions are confirmed ready. Backfill alone does not switch reads.
func (s *Attendance) ActivateQueries(ctx context.Context, progress QueryMigration) error {
	if progress.Version != 1 || progress.ChannelID != s.ChannelID || !progress.Apply || !progress.Complete {
		return errors.New("activation requires a completed apply checkpoint for this channel")
	}
	_, err := s.queryState().Set(ctx, map[string]any{"ready": true, "version": 1, "completedAt": time.Now().UTC()})
	return err
}
