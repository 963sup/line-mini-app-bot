package firestore

import (
	"context"
	"crypto/sha256"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"reflect"
	"testing"
	"time"

	db "cloud.google.com/go/firestore"
	"github.com/963sup/line-mini-app-bot/internal/attendance"
	"github.com/963sup/line-mini-app-bot/internal/attendance/worktime"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func legacyHash(c attendance.Correction) string {
	c.ID, c.UserKey, c.CreatedAt = "", "", time.Time{}
	c.Start, c.End = c.Start.UTC(), c.End.UTC()
	c.Breaks = append([]worktime.Interval{}, c.Breaks...)
	for i := range c.Breaks {
		c.Breaks[i].Start, c.Breaks[i].End = c.Breaks[i].Start.UTC(), c.Breaks[i].End.UTC()
	}
	body, _ := json.Marshal(struct {
		Version int
		Payload legacyCorrection
	}{1, legacyCorrectionFrom(c)})
	return fmt.Sprintf("%x", sha256.Sum256(body))
}

func TestReceiptHashRetainsLegacyEncoding(t *testing.T) {
	start := time.Date(2026, 11, 1, 8, 0, 0, 0, time.FixedZone("Taipei", 8*3600))
	c := attendance.Correction{ID: "ignored", UserKey: "private", CreatedAt: start, GroupID: "g", RecordID: "r", Reason: "reason", Start: start, End: start.Add(time.Hour), RestConfirmed: true, ExpectedRevision: 7}
	for _, breaks := range [][]worktime.Interval{nil, {}, {{Start: start.Add(time.Minute), End: start.Add(2 * time.Minute)}}} {
		c.Breaks = breaks
		got, err := correctionPayloadHash(c)
		if err != nil {
			t.Fatal(err)
		}
		if got != legacyHash(c) {
			t.Fatal("persisted receipt hash changed")
		}
	}
}

func TestTechnicalErrorsStayInsideStorage(t *testing.T) {
	raw := status.Error(codes.NotFound, "private document details")
	if !errors.Is(translateError(raw, attendance.ErrCorrection), attendance.ErrCorrection) {
		t.Fatal("missing correction not translated")
	}
	got := translateError(raw, nil)
	if !errors.Is(got, attendance.ErrUnavailable) || status.Code(got) == codes.NotFound || got.Error() == raw.Error() {
		t.Fatal("technical error escaped adapter")
	}
	if translateError(status.Error(codes.DeadlineExceeded, "timeout"), nil) != context.DeadlineExceeded {
		t.Fatal("gRPC timeout was not translated")
	}
	if translateError(status.Error(codes.Canceled, "canceled"), nil) != context.Canceled {
		t.Fatal("gRPC cancellation was not translated")
	}
	if translateError(context.Canceled, nil) != context.Canceled {
		t.Fatal("context cancellation lost")
	}
}

func TestFirestoreDocumentsRemainCompatible(t *testing.T) {
	if os.Getenv("FIRESTORE_EMULATOR_HOST") == "" {
		t.Skip("requires local Firestore emulator")
	}
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	client, err := db.NewClient(ctx, "boundary-local-test")
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		if err := client.Close(); err != nil {
			t.Error(err)
		}
	}()
	now := time.Date(2026, 11, 1, 8, 0, 0, 123000000, time.UTC)
	record := attendance.Record{RecordID: "r1", GroupID: "g", ClockInAt: now, IsOpen: true, ClockInPosition: &attendance.Position{Latitude: 25.1234567890123, Longitude: 121.9876543210123, CapturedAt: now, AccuracyMeters: 1}, Breaks: []worktime.Interval{{Start: now, End: now.Add(time.Minute)}}, SourceEventIDs: []string{"event"}, SchemaVersion: 3}
	oldRecord := legacyRecord{RecordID: record.RecordID, GroupID: record.GroupID, ClockInAt: record.ClockInAt, IsOpen: record.IsOpen, ClockInPosition: &legacyPosition{Latitude: record.ClockInPosition.Latitude, Longitude: record.ClockInPosition.Longitude, CapturedAt: now, AccuracyMeters: 1}, Breaks: record.Breaks, SourceEventIDs: record.SourceEventIDs, SchemaVersion: 3}
	correction := attendance.Correction{ID: "c1", UserKey: documentKey("u"), GroupID: "g", RecordID: "r1", Start: now, End: now.Add(time.Hour), Reason: "reason", RestConfirmed: true, CreatedAt: now}
	request := requestDocument{GroupID: "g", UserID: documentKey("u"), EventID: "e", Action: "clock-in", EventTime: now.UnixMilli(), ExpiresAt: now.Add(time.Hour)}
	cases := []struct {
		name     string
		old, new any
	}{
		{"record", oldRecord, recordDocumentFrom(record)},
		{"result", legacyResult{Code: "clocked_in", Record: oldRecord}, resultDocumentFrom(attendance.Result{Code: "clocked_in", Record: record})},
		{"request", legacyRequest(request), request},
		{"correction", legacyCorrectionFrom(correction), correctionDocumentFrom(correction)},
	}
	root := client.Collection("codecBoundary").Doc(fmt.Sprint(time.Now().UnixNano())).Collection("documents")
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			left, right := root.Doc(tc.name+"-old"), root.Doc(tc.name+"-new")
			if _, err := left.Set(ctx, tc.old); err != nil {
				t.Fatal(err)
			}
			if _, err := right.Set(ctx, tc.new); err != nil {
				t.Fatal(err)
			}
			a, err := left.Get(ctx)
			if err != nil {
				t.Fatal(err)
			}
			b, err := right.Get(ctx)
			if err != nil {
				t.Fatal(err)
			}
			if !reflect.DeepEqual(a.Data(), b.Data()) {
				t.Fatalf("Firestore format changed: %#v %#v", a.Data(), b.Data())
			}
			switch tc.name {
			case "record":
				var got attendance.Record
				if err := decodeRecord(a, &got); err != nil {
					t.Fatal(err)
				}
				if !reflect.DeepEqual(got, record) {
					t.Fatalf("legacy record decode changed: %#v", got)
				}
			case "result":
				var got attendance.Result
				if err := decodeResult(a, &got); err != nil || got.Record.RecordID != "r1" {
					t.Fatal("legacy receipt decode failed", err)
				}
			case "correction":
				var got attendance.Correction
				if err := decodeCorrection(a, &got); err != nil || !reflect.DeepEqual(got, correction) {
					t.Fatal("legacy correction decode failed", err)
				}
			}
		})
	}
	// A receipt written before this refactor must replay without reading a revised/deleted shift.
	store := &Attendance{Client: client, ChannelID: fmt.Sprint("legacy-replay-", time.Now().UnixNano())}
	key := "00000000-0000-4000-8000-000000000001"
	expected := documentKey("correction", "u", key)
	_, err = store.user("u").Collection("attendanceCorrectionReceipts").Doc(documentKey(key)).Set(ctx, correctionReceipt{Version: 1, PayloadHash: legacyHash(correction), CorrectionID: expected})
	if err != nil {
		t.Fatal(err)
	}
	id, err := store.ProposeCorrectionOnce(ctx, "u", key, correction, now)
	if err != nil || id != expected {
		t.Fatalf("old receipt no longer replays: %s %v", id, err)
	}
}

func legacyCorrectionFrom(c attendance.Correction) legacyCorrection {
	return legacyCorrection{ID: c.ID, UserKey: c.UserKey, GroupID: c.GroupID, RecordID: c.RecordID, ExpectedEventSequence: c.ExpectedEventSequence, ExpectedRevision: c.ExpectedRevision, Reason: c.Reason, Start: c.Start, End: c.End, Breaks: c.Breaks, RestConfirmed: c.RestConfirmed, CreatedAt: c.CreatedAt}
}
