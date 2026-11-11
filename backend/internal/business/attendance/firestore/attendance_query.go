package firestore

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"sort"
	"strings"
	"time"

	db "cloud.google.com/go/firestore"
	recordmodel "github.com/963sup/line-mini-app-bot/internal/business/attendance/records"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

const queryBatchSize = 200

// QueryMode is explicit so callers can disclose historical-read costs during rollout.
func (s *Attendance) QueryMode(ctx context.Context) (string, error) {
	snap, err := s.queryState().Get(ctx)
	if status.Code(err) == codes.NotFound {
		return "legacy", nil
	}
	if err != nil {
		return "", err
	}
	ready, ok := snap.Data()["ready"].(bool)
	if ok && ready {
		return "indexed", nil
	}
	return "legacy", nil
}

func (s *Attendance) queryState() *db.DocumentRef {
	return s.Client.Collection("channels").Doc(s.ChannelID).Collection("attendanceQueryState").Doc("v1")
}

type pageCursor struct {
	Version int       `json:"v"`
	Scope   string    `json:"s"`
	Time    time.Time `json:"t,omitempty"`
	End     time.Time `json:"e,omitempty"`
	ID      string    `json:"i"`
}

func (s *Attendance) cursorScope(kind, user, group string, start time.Time) string {
	return documentKey("attendance-query-v1-start-end-id-desc", s.ChannelID, user, kind, group, start.UTC().Format(time.RFC3339Nano))
}

func decodeCursor(value, scope string) (pageCursor, error) {
	if value == "" {
		return pageCursor{}, nil
	}
	if len(value) > 1024 {
		return pageCursor{}, recordmodel.ErrInvalidCursor
	}
	b, err := base64.RawURLEncoding.DecodeString(value)
	if err != nil {
		return pageCursor{}, recordmodel.ErrInvalidCursor
	}
	var c pageCursor
	if json.Unmarshal(b, &c) != nil || c.Version != 1 || c.Scope != scope || !validSegment(c.ID) {
		return pageCursor{}, recordmodel.ErrInvalidCursor
	}
	return c, nil
}

func encodeCursor(scope string, at time.Time, id string) string {
	return encodeRecordCursor(scope, at, time.Time{}, id)
}

func encodeRecordCursor(scope string, at, end time.Time, id string) string {
	b, _ := json.Marshal(pageCursor{Version: 1, Scope: scope, Time: at.UTC(), End: end.UTC(), ID: id})
	return base64.RawURLEncoding.EncodeToString(b)
}

func validSegment(id string) bool {
	return id != "" && len(id) <= 1500 && id != "." && id != ".." && !strings.Contains(id, "/")
}
func validPageSize(size int) bool { return size >= 1 && size <= 100 }

func readQuery(ctx context.Context, q db.Query) ([]*db.DocumentSnapshot, error) {
	docs := []*db.DocumentSnapshot{}
	for {
		batch, err := q.Limit(queryBatchSize).Documents(ctx).GetAll()
		if err != nil {
			return nil, err
		}
		docs = append(docs, batch...)
		if len(batch) < queryBatchSize {
			return docs, nil
		}
		q = q.StartAfter(batch[len(batch)-1])
	}
}

func decodeAttendance(docs []*db.DocumentSnapshot) ([]recordmodel.Record, error) {
	records := make([]recordmodel.Record, 0, len(docs))
	for _, doc := range docs {
		var record recordmodel.Record
		if err := decodeRecord(doc, &record); err != nil {
			return nil, err
		}
		record.RecordID = doc.Ref.ID
		records = append(records, record)
	}
	return records, nil
}

func overlapsMonth(r recordmodel.Record, start time.Time) bool {
	return r.ClockInAt.Before(start.AddDate(0, 1, 0)) && (r.ClockOutAt.IsZero() || r.ClockOutAt.After(start))
}

func sortRecords(records []recordmodel.Record) {
	sort.Slice(records, func(i, j int) bool {
		if records[i].ClockInAt.Equal(records[j].ClockInAt) {
			if !records[i].ClockOutAt.Equal(records[j].ClockOutAt) {
				return records[i].ClockOutAt.After(records[j].ClockOutAt)
			}
			return records[i].RecordID > records[j].RecordID
		}
		return records[i].ClockInAt.After(records[j].ClockInAt)
	})
}

// monthQueries keep open shifts explicit: absent clockOutAt fields cannot satisfy inequalities.
// Closed queries explicitly order both inequality fields before the document key.
// The merged order is start DESC, end DESC (open shifts last), key DESC.
func (s *Attendance) monthQueries(user, group string, start time.Time, c pageCursor) []db.Query {
	base := s.user(user).Collection("attendance").Where("groupID", "==", group).Where("clockInAt", "<", start.AddDate(0, 1, 0))
	closed := base.Where("clockOutAt", ">", start).OrderBy("clockInAt", db.Desc).OrderBy("clockOutAt", db.Desc).OrderBy(db.DocumentID, db.Desc)
	open := base.Where("isOpen", "==", true).OrderBy("clockInAt", db.Desc).OrderBy(db.DocumentID, db.Desc)
	if c.ID != "" {
		if c.End.IsZero() {
			// At a tied start every closed shift precedes every open shift.
			closed = closed.Where("clockInAt", "<", c.Time)
			open = open.StartAfter(c.Time, c.ID)
		} else {
			closed = closed.StartAfter(c.Time, c.End, c.ID)
			open = open.Where("clockInAt", "<=", c.Time)
		}
	}
	return []db.Query{closed, open}
}

func (s *Attendance) monthCandidates(ctx context.Context, user, group string, start time.Time) ([]recordmodel.Record, error) {
	records := []recordmodel.Record{}
	seen := map[string]bool{}
	for _, q := range s.monthQueries(user, group, start, pageCursor{}) {
		docs, err := readQuery(ctx, q)
		if err != nil {
			return nil, err
		}
		batch, err := decodeAttendance(docs)
		if err != nil {
			return nil, err
		}
		for _, r := range batch {
			if !seen[r.RecordID] {
				records = append(records, r)
				seen[r.RecordID] = true
			}
		}
	}
	sortRecords(records)
	return records, nil
}

func (s *Attendance) GroupsPage(ctx context.Context, user, cursor string, size int) (_ recordmodel.GroupPage, err error) {
	defer func() { err = translateError(err, nil) }()
	page := recordmodel.GroupPage{Items: []string{}}
	if !validPageSize(size) {
		return page, recordmodel.ErrInvalidCursor
	}
	scope := s.cursorScope("groups", user, "", time.Time{})
	c, err := decodeCursor(cursor, scope)
	if err != nil {
		return page, err
	}
	mode, err := s.QueryMode(ctx)
	page.QueryMode = mode
	if err != nil {
		return page, err
	}
	groups := []string{}
	if mode == "legacy" {
		groups, err = s.legacyGroups(ctx, user)
		if err != nil {
			return page, err
		}
		at := sort.SearchStrings(groups, c.ID)
		if c.ID != "" {
			for at < len(groups) && groups[at] <= c.ID {
				at++
			}
		}
		groups = groups[at:]
	} else {
		q := s.user(user).Collection("attendanceGroups").OrderBy("groupID", db.Asc)
		if c.ID != "" {
			q = q.StartAfter(c.ID)
		}
		docs, e := q.Limit(size + 1).Documents(ctx).GetAll()
		if e != nil {
			return page, e
		}
		for _, doc := range docs {
			group, ok := doc.Data()["groupID"].(string)
			if !ok || !validSegment(group) {
				return page, recordmodel.ErrInvalidCursor
			}
			groups = append(groups, group)
		}
	}
	if len(groups) > size {
		page.NextCursor = encodeCursor(scope, time.Time{}, groups[size-1])
		groups = groups[:size]
	}
	page.Items = groups
	return page, nil
}

func (s *Attendance) RecordsPage(ctx context.Context, user, group string, start time.Time, cursor string, size int) (_ recordmodel.RecordPage, err error) {
	defer func() { err = translateError(err, nil) }()
	page := recordmodel.RecordPage{Items: []recordmodel.Record{}}
	if !validPageSize(size) {
		return page, recordmodel.ErrInvalidCursor
	}
	scope := s.cursorScope("records", user, group, start)
	c, err := decodeCursor(cursor, scope)
	if err != nil || (cursor != "" && c.Time.IsZero()) {
		return page, recordmodel.ErrInvalidCursor
	}
	mode, err := s.QueryMode(ctx)
	page.QueryMode = mode
	if err != nil {
		return page, err
	}
	records := []recordmodel.Record{}
	if mode == "legacy" {
		records, err = s.legacyGroupRecords(ctx, user, group)
		if err != nil {
			return page, err
		}
		filtered := []recordmodel.Record{}
		for _, r := range records {
			after := c.ID == "" || r.ClockInAt.Before(c.Time) || (r.ClockInAt.Equal(c.Time) && (r.ClockOutAt.Before(c.End) || (r.ClockOutAt.Equal(c.End) && r.RecordID < c.ID)))
			if after && overlapsMonth(r, start) {
				filtered = append(filtered, r)
			}
		}
		records = filtered
	} else {
		seen := map[string]bool{}
		for _, q := range s.monthQueries(user, group, start, c) {
			docs, e := q.Limit(size + 1).Documents(ctx).GetAll()
			if e != nil {
				return page, e
			}
			batch, e := decodeAttendance(docs)
			if e != nil {
				return page, e
			}
			for _, r := range batch {
				if !seen[r.RecordID] {
					records = append(records, r)
					seen[r.RecordID] = true
				}
			}
		}
	}
	sortRecords(records)
	if len(records) > size {
		last := records[size-1]
		page.NextCursor = encodeRecordCursor(scope, last.ClockInAt, last.ClockOutAt, last.RecordID)
		records = records[:size]
	}
	page.Items, err = s.verifyAndRebuildRecords(ctx, user, group, records)
	return page, err
}

func (s *Attendance) Record(ctx context.Context, user, group, id string) (_ recordmodel.Record, err error) {
	defer func() { err = translateError(err, nil) }()
	if !validSegment(id) {
		return recordmodel.Record{}, recordmodel.ErrRecordNotFound
	}
	doc, err := s.user(user).Collection("attendance").Doc(id).Get(ctx)
	if status.Code(err) == codes.NotFound {
		return recordmodel.Record{}, recordmodel.ErrRecordNotFound
	}
	if err != nil {
		return recordmodel.Record{}, err
	}
	records, err := decodeAttendance([]*db.DocumentSnapshot{doc})
	if err != nil {
		return recordmodel.Record{}, err
	}
	if records[0].GroupID != group {
		return recordmodel.Record{}, recordmodel.ErrRecordNotFound
	}
	records, err = s.verifyAndRebuildRecords(ctx, user, group, records)
	if err != nil {
		return recordmodel.Record{}, err
	}
	return records[0], nil
}
