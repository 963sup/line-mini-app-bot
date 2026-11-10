package firestore

import (
	"context"
	"errors"
	"fmt"
	"os"
	"testing"
	"time"

	db "cloud.google.com/go/firestore"
	"github.com/963sup/line-mini-app-bot/internal/attendance"
	"github.com/963sup/line-mini-app-bot/internal/attendance/worktime"
)

// Never runs against a production Firestore project.
func TestAttendanceCorrectionRoundTrip(t *testing.T) {
	if os.Getenv("FIRESTORE_EMULATOR_HOST") == "" {
		t.Skip("requires Firestore emulator; integration not verified")
	}
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	client, err := db.NewClient(ctx, "attendance-local-test")
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		if err := client.Close(); err != nil {
			t.Errorf("close Firestore client: %v", err)
		}
	}()
	s := &Attendance{Client: client, ChannelID: fmt.Sprintf("closure-%d", time.Now().UnixNano())}
	zone := time.FixedZone("Taipei", 28800)
	start := time.Date(2026, 11, 2, 9, 0, 0, 0, zone)
	if _, err := s.Punch(ctx, "group", "employee", "in", "clock-in", start, 1); err != nil {
		t.Fatal(err)
	}
	month := time.Date(2026, 11, 1, 0, 0, 0, 0, zone)
	records, err := s.MonthGroup(ctx, "employee", "group", month)
	if err != nil || len(records) != 1 {
		t.Fatalf("records=%v err=%v", records, err)
	}
	record := records[0]
	c := attendance.Correction{RecordID: record.RecordID, GroupID: "group", ExpectedEventSequence: record.EventSequence, Reason: "漏下班及確認休息", Start: start, End: start.Add(8 * time.Hour), RestConfirmed: true}
	key := "f6ed3cb6-5dc6-4a43-ae34-984d673299bf"
	type submission struct {
		id  string
		err error
	}
	submissions := make(chan submission, 2)
	for i := 0; i < 2; i++ {
		go func() {
			id, err := s.ProposeCorrectionOnce(ctx, "employee", key, c, start.Add(9*time.Hour))
			submissions <- submission{id, err}
		}()
	}
	first, second := <-submissions, <-submissions
	id := first.id
	if first.err != nil || second.err != nil || first.id == "" || second.id != first.id {
		t.Fatalf("concurrent submission: %+v %+v", first, second)
	}
	// Timestamp offsets and empty break representations are the same payload.
	replay := c
	replay.Start, replay.End = c.Start.UTC(), c.End.UTC()
	replay.Breaks = []worktime.Interval{}
	if again, err := s.ProposeCorrectionOnce(ctx, "employee", key, replay, start.Add(10*time.Hour)); err != nil || again != id {
		t.Fatalf("submission replay: id=%s err=%v", again, err)
	}
	changedPayload := c
	changedPayload.Reason = "different proposal"
	if _, err := s.ProposeCorrectionOnce(ctx, "employee", key, changedPayload, start.Add(10*time.Hour)); !errors.Is(err, attendance.ErrIdempotencyConflict) {
		t.Fatalf("changed payload reused key: %v", err)
	}
	if _, err := s.ProposeCorrectionOnce(ctx, "stranger", key, c, start.Add(10*time.Hour)); err == nil {
		t.Fatal("another actor replayed employee submission")
	}
	if state, err := s.CorrectionStatus(ctx, "employee", id); err != nil || state != "pending" {
		t.Fatalf("%s %v", state, err)
	}
	if _, err := s.CorrectionStatus(ctx, "stranger", id); err == nil {
		t.Fatal("other employee read correction")
	}
	roles := client.Collection("channels").Doc(s.ChannelID).Collection("attendanceReviewers")
	for _, reviewer := range []string{"reviewer", "employee"} {
		if _, err := roles.Doc(documentKey(reviewer)).Set(ctx, map[string]any{"groups": []string{"group"}}); err != nil {
			t.Fatal(err)
		}
	}
	if err := s.ApproveCorrection(ctx, "employee", id, start.Add(10*time.Hour)); err == nil {
		t.Fatal("self approval accepted")
	}
	if err := s.ApproveCorrection(ctx, "reviewer", id, start.Add(10*time.Hour)); err != nil {
		t.Fatal(err)
	}
	if err := s.ApproveCorrection(ctx, "reviewer", id, start.Add(11*time.Hour)); err != nil {
		t.Fatal("duplicate approval", err)
	}
	// A committed submission remains replayable after approval changes revision.
	if again, err := s.ProposeCorrectionOnce(ctx, "employee", key, c, start.Add(12*time.Hour)); err != nil || again != id {
		t.Fatalf("approved submission replay: id=%s err=%v", again, err)
	}
	if state, err := s.CorrectionStatus(ctx, "employee", id); err != nil || state != "approved" {
		t.Fatalf("%s %v", state, err)
	}
	records, err = s.MonthGroup(ctx, "employee", "group", month)
	if err != nil {
		t.Fatal(err)
	}
	report := attendance.CalculateMonthlyAttendance(records, month)
	if report.Calculation.Minutes != 480 || report.Calculation.Pending != 0 || records[0].Revision != 1 {
		t.Fatalf("%+v", report)
	}
	// A corrupt approval must fail in both status and the monthly report.
	audit := client.Collection("channels").Doc(s.ChannelID).Collection("attendanceCorrections").Doc(id)
	if _, err := audit.Update(ctx, []db.Update{{Path: "proposal.reason", Value: ""}}); err != nil {
		t.Fatal(err)
	}
	if state, err := s.CorrectionStatus(ctx, "employee", id); err != nil || state != "invalid" {
		t.Fatalf("corrupt status=%s %v", state, err)
	}
	records, err = s.MonthGroup(ctx, "employee", "group", month)
	if err != nil {
		t.Fatal(err)
	}
	if report := attendance.CalculateMonthlyAttendance(records, month); report.Calculation.Minutes != 0 || report.Calculation.Pending != 1 {
		t.Fatalf("corrupt approval counted: %+v", report)
	}
	// Extend the existing round trip through the staged legacy-to-indexed rollout.
	// These fixtures deliberately have missing parent user documents and legacy query metadata.
	sameTime := month.Add(10 * time.Hour)
	fixtures := map[string]attendance.Record{
		"legacy-a":          {GroupID: "group", SchemaVersion: 2, ClockInAt: sameTime, ClockOutAt: sameTime.Add(time.Hour), RestConfirmed: true},
		"legacy-b":          {GroupID: "group", SchemaVersion: 2, ClockInAt: sameTime, ClockOutAt: sameTime.Add(2 * time.Hour), RestConfirmed: true},
		"legacy-c":          {GroupID: "group", SchemaVersion: 2, ClockInAt: sameTime, ClockOutAt: sameTime.Add(time.Hour), RestConfirmed: true},
		"same-time-open":    {GroupID: "group", SchemaVersion: 2, ClockInAt: sameTime},
		"cross-month":       {GroupID: "group", SchemaVersion: 2, ClockInAt: month.Add(-time.Hour), ClockOutAt: month.Add(time.Hour), RestConfirmed: true},
		"open-before-month": {GroupID: "group", SchemaVersion: 2, ClockInAt: month.Add(-2 * time.Hour)},
		"old-closed":        {GroupID: "group", SchemaVersion: 2, ClockInAt: month.AddDate(-1, 0, 0), ClockOutAt: month.AddDate(-1, 0, 0).Add(time.Hour)},
		"other-group":       {GroupID: "group-b", SchemaVersion: 2, ClockInAt: sameTime, ClockOutAt: sameTime.Add(time.Hour)},
	}
	for id, r := range fixtures {
		if _, err := s.user("employee").Collection("attendance").Doc(id).Set(ctx, recordDocumentFrom(r)); err != nil {
			t.Fatal(err)
		}
	}
	var dry QueryMigration
	for !dry.Complete {
		dry, err = s.MigrateQueryBatch(ctx, dry, false)
		if err != nil {
			t.Fatal(err)
		}
	}
	mode, err := s.QueryMode(ctx)
	if err != nil || mode != "legacy" {
		t.Fatalf("dry-run activated reads: %s %v", mode, err)
	}
	var applied QueryMigration
	for !applied.Complete {
		applied, err = s.MigrateQueryBatch(ctx, applied, true)
		if err != nil {
			t.Fatal(err)
		}
	}
	mode, err = s.QueryMode(ctx)
	if err != nil || mode != "legacy" {
		t.Fatalf("backfill activated reads without confirmation: %s %v", mode, err)
	}
	legacyIDs := []string{}
	legacyCursor := ""
	for {
		page, e := s.RecordsPage(ctx, "employee", "group", month, legacyCursor, 1)
		if e != nil {
			t.Fatal(e)
		}
		for _, r := range page.Items {
			legacyIDs = append(legacyIDs, r.RecordID)
		}
		if page.NextCursor == "" {
			break
		}
		legacyCursor = page.NextCursor
	}
	if err := s.ActivateQueries(ctx, applied); err != nil {
		t.Fatal(err)
	}
	page, err := s.GroupsPage(ctx, "employee", "", 1)
	if err != nil || page.QueryMode != "indexed" || len(page.Items) != 1 || page.Items[0] != "group" || page.NextCursor == "" {
		t.Fatalf("groups first page: %+v %v", page, err)
	}
	next, err := s.GroupsPage(ctx, "employee", page.NextCursor, 1)
	if err != nil || len(next.Items) != 1 || next.Items[0] != "group-b" || next.NextCursor != "" {
		t.Fatalf("groups next page: %+v %v", next, err)
	}
	seen := map[string]bool{}
	indexedIDs := []string{}
	cursor := ""
	for {
		page, err := s.RecordsPage(ctx, "employee", "group", month, cursor, 1)
		if err != nil {
			t.Fatal(err)
		}
		if len(page.Items) != 1 || seen[page.Items[0].RecordID] {
			t.Fatalf("repeated or missing record: %+v", page)
		}
		seen[page.Items[0].RecordID] = true
		indexedIDs = append(indexedIDs, page.Items[0].RecordID)
		if page.NextCursor == "" {
			break
		}
		cursor = page.NextCursor
	}
	if len(seen) != 7 || !seen["legacy-a"] || !seen["legacy-b"] || !seen["cross-month"] || !seen["open-before-month"] || seen["old-closed"] {
		t.Fatalf("incorrect month page set: %v", seen)
	}
	if _, err := s.RecordsPage(ctx, "another-employee", "group", month, cursor, 1); err != attendance.ErrInvalidCursor {
		t.Fatalf("cross-user cursor accepted: %v", err)
	}
	if fmt.Sprint(legacyIDs) != fmt.Sprint(indexedIDs) {
		t.Fatalf("legacy/indexed order mismatch: %v / %v", legacyIDs, indexedIDs)
	}
	indexed, err := s.MonthGroup(ctx, "employee", "group", month)
	if err != nil || len(indexed) != len(seen) {
		t.Fatalf("indexed month differs from pages: %d %v", len(indexed), err)
	}
	recent, err := s.RecentGroup(ctx, "employee", "group")
	if err != nil || len(recent) != 5 {
		t.Fatalf("indexed recent: %d %v", len(recent), err)
	}
	if _, err := s.Record(ctx, "employee", "group-b", "legacy-a"); err != attendance.ErrRecordNotFound {
		t.Fatalf("cross-group detail accepted: %v", err)
	}
}

func TestLaunchExchangeIsStableAndScoped(t *testing.T) {
	if os.Getenv("FIRESTORE_EMULATOR_HOST") == "" {
		t.Skip("requires Firestore emulator")
	}
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	client, err := db.NewClient(ctx, "attendance-local-test")
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		if err := client.Close(); err != nil {
			t.Errorf("close Firestore client: %v", err)
		}
	}()
	s := &Attendance{Client: client, ChannelID: fmt.Sprintf("launch-%d", time.Now().UnixNano())}
	token, err := s.CreateLaunch(ctx, "group", "clock-in")
	if err != nil {
		t.Fatal(err)
	}
	group, err := s.LaunchGroup(ctx, token)
	if err != nil || group != "group" {
		t.Fatal("launch must stay group bound")
	}
	first, err := s.ExchangeLaunch(ctx, token, "user-a")
	if err != nil {
		t.Fatal(err)
	}
	second, err := s.ExchangeLaunch(ctx, token, "user-a")
	if err != nil || second != first {
		t.Fatal("same user/card must reuse request")
	}
	other, err := s.ExchangeLaunch(ctx, token, "user-b")
	if err != nil || other == first {
		t.Fatal("each member needs separate request")
	}
	snap, err := s.requests().Doc(first).Get(ctx)
	if err != nil {
		t.Fatal(err)
	}
	var request requestDocument
	if err = snap.DataTo(&request); err != nil {
		t.Fatal(err)
	}
	if request.GroupID != "group" || request.Action != "clock-in" || request.UserID != documentKey("user-a") {
		t.Fatal("request binding changed")
	}
}
