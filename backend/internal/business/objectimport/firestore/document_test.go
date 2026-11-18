package firestore

import (
	"encoding/json"
	"reflect"
	"strings"
	"testing"
	"time"

	feature "github.com/963sup/line-mini-app-bot/internal/business/objectimport"
)

func TestV1SnapshotRoundTripAndSize(t *testing.T) {
	now := time.Now().UTC().Truncate(time.Microsecond)
	j := feature.Job{ID: strings.Repeat("a", 32), RepositoryID: strings.Repeat("b", 32), CreatorID: "A", Status: "completed", Version: 4, ConfirmedFromReviewVersion: 2, SourceDocument: "source", SourceRevision: "v0", SourceHash: strings.Repeat("c", 64), CreatedAt: now, UpdatedAt: now, ReviewedAt: &now, Rows: []feature.Row{{ID: strings.Repeat("d", 32), Fields: feature.Fields{Name: "項", Description: "完整內容", Type: "type", Quantity: "11", Unit: "SET", MaterialCode: "code"}, SourceItemNo: "10", Evidence: []feature.Evidence{{Page: 2, Quote: "literal"}}, ResultObjectID: strings.Repeat("e", 32)}}}
	stored := toStored(j)
	b, e := json.Marshal(stored)
	if e != nil {
		t.Fatal(e)
	}
	var parsed storedJob
	if e = json.Unmarshal(b, &parsed); e != nil {
		t.Fatal(e)
	}
	if got := fromStored(parsed); !reflect.DeepEqual(got, j) {
		t.Fatalf("lost v1 fields: %+v", got)
	}
	domain, _ := json.Marshal(j)
	if len(b) > len(domain) {
		t.Fatal("wire grew beyond conservative domain budget")
	}
	if _, e = encode(j); e != nil {
		t.Fatal(e)
	}
	j.ReviewedAt = nil
	j.Status = "draft"
	j.ConfirmedFromReviewVersion = 0
	j.Rows[0].ResultObjectID = ""
	stored = toStored(j)
	b, _ = json.Marshal(stored)
	if strings.Contains(string(b), "reviewedAt") {
		t.Fatal("draft serialized absent review time")
	}
}
