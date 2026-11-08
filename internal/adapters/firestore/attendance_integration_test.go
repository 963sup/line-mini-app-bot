package firestore

import (
	db "cloud.google.com/go/firestore"
	"context"
	"fmt"
	"github.com/963sup/line-mini-app-bot/internal/attendance"
	"os"
	"testing"
	"time"
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
	defer client.Close()
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
	id, err := s.ProposeCorrection(ctx, "employee", c, start.Add(9*time.Hour))
	if err != nil {
		t.Fatal(err)
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
}
