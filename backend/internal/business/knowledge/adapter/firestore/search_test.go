package firestore

import (
	"context"
	"os"
	"testing"
	"time"

	db "cloud.google.com/go/firestore"
	"github.com/963sup/line-mini-app-bot/internal/business/knowledge/application/port"
	"github.com/963sup/line-mini-app-bot/internal/platform/firestore/key"
	"github.com/google/uuid"
)

func TestSearchEvidenceChecksPublicationAndOrganization(t *testing.T) {
	if os.Getenv("FIRESTORE_EMULATOR_HOST") == "" {
		t.Skip("requires Firestore emulator")
	}
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()
	client, err := db.NewClient(ctx, "knowledge-local-test")
	if err != nil {
		t.Fatal(err)
	}
	defer client.Close()

	reader := Reader{DB: client, Channel: uuid.NewString()}
	const template = "template-v1"
	seed := func(actor string) string {
		t.Helper()
		id := uuid.NewString()
		organization := client.Collection("channels").Doc(reader.Channel).Collection("organizations").Doc(id)
		set := func(ref *db.DocumentRef, value any) {
			t.Helper()
			if _, err := ref.Set(ctx, value); err != nil {
				t.Fatal(err)
			}
		}
		set(organization, map[string]any{"ownerUserId": actor})
		// Publish the near and far entries only. The unpublished ref ranks first.
		set(organization.Collection("publishedKnowledge").Doc(firestorekey.Digest([]byte(template))), map[string]any{
			"templateVersion": template,
			"rules": []port.PublishedRule{
				{ID: "near", Version: "v2", Source: "p.1", Text: "Near rule"},
				{ID: "far", Version: "v1", Source: "p.2", Text: "Far rule"},
			},
		})
		for _, ref := range []struct {
			id, version string
			distance    float64
		}{{"near", "v1", 0}, {"near", "v2", 1}, {"far", "v1", 2}} {
			set(organization.Collection("publishedKnowledgeVectors").Doc(ref.id+"-"+ref.version), map[string]any{
				"templateVersion": template, "ruleId": ref.id, "ruleVersion": ref.version,
				"embedding": db.Vector64{ref.distance, 0},
			})
		}
		return id
	}
	own, other := seed("member"), seed("other")
	q := port.EvidenceSearch{Actor: "member", Organization: own, Vector: []float64{0, 0}, Limit: 3}
	if rules, err := reader.SearchEvidence(ctx, q); err != nil || len(rules) != 2 || rules[0].ID != "near" || rules[0].Version != "v2" || rules[1].ID != "far" {
		t.Fatalf("nearest published rules: %+v, %v", rules, err)
	}
	q.Organization = other
	if rules, err := reader.SearchEvidence(ctx, q); err == nil || len(rules) != 0 {
		t.Fatalf("cross-organization result: %+v, %v", rules, err)
	}
	q.Organization = own
	if _, err := client.Collection("channels").Doc(reader.Channel).Collection("organizations").Doc(own).
		Collection("publishedKnowledge").Doc(firestorekey.Digest([]byte(template))).Delete(ctx); err != nil {
		t.Fatal(err)
	}
	if rules, err := reader.SearchEvidence(ctx, q); err != nil || len(rules) != 0 {
		t.Fatalf("withdrawn rules: %+v, %v", rules, err)
	}
}
