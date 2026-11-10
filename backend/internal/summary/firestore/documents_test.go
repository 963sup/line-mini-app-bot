package firestore

import (
	"context"
	"fmt"
	"os"
	"reflect"
	"testing"
	"time"

	db "cloud.google.com/go/firestore"
	"github.com/963sup/line-mini-app-bot/internal/summary"
)

func TestSummaryDocumentBoundaryCompatibility(t *testing.T) {
	if os.Getenv("FIRESTORE_EMULATOR_HOST") == "" {
		t.Skip("requires local Firestore emulator")
	}
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()
	client, err := db.NewClient(ctx, "summary-boundary-local-test")
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		if err := client.Close(); err != nil {
			t.Error(err)
		}
	}()
	now := time.Date(2026, 11, 1, 8, 0, 0, 0, time.UTC)
	message := summary.Message{ID: "m", EventID: "event", UserID: "actor", Text: "text", OccurredAt: now, ExpireAt: now.Add(time.Hour), Epoch: 2, EditedAt: 3}
	job := summary.Job{ID: "j", GroupID: "g", Kind: "summary", RetryKey: "retry", Status: "pending", Epoch: 2, Revision: 3, Cutoff: now, ExpireAt: now.Add(time.Hour), Sources: []string{"m"}}
	root := client.Collection("boundary").Doc(fmt.Sprint(time.Now().UnixNano())).Collection("docs")
	for _, tc := range []struct {
		name     string
		old, new any
	}{{"message", message, messageDocument(message)}, {"job", job, jobDocument(job)}} {
		left, right := root.Doc(tc.name+"old"), root.Doc(tc.name+"new")
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
			t.Fatal("summary document format changed")
		}
		if _, ok := b.Data()["ID"]; !ok {
			t.Fatal("legacy upper-case ID changed")
		}
		if tc.name == "message" {
			var got summary.Message
			if err := decodeMessage(a, &got); err != nil || !reflect.DeepEqual(got, message) {
				t.Fatal("legacy message read failed", err)
			}
		} else {
			var got summary.Job
			if err := decodeJob(a, &got); err != nil || !reflect.DeepEqual(got, job) {
				t.Fatal("legacy job read failed", err)
			}
		}
	}
}
