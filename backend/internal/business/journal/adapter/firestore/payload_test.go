package firestore

import (
	"encoding/json"
	"reflect"
	"testing"
	"time"

	journalentity "github.com/963sup/line-mini-app-bot/internal/business/journal/domain/entity"
	firestorekey "github.com/963sup/line-mini-app-bot/internal/platform/firestore/key"
)

// Golden bytes and hashes were captured from the pre-refactor domain encoder.
func TestPayloadCompatibility(t *testing.T) {
	date := time.Date(2026, 11, 23, 0, 0, 0, 0, time.UTC)
	tests := []struct {
		name              string
		draft             journalentity.Draft
		json, hash, entry string
	}{
		{name: "legacy", draft: journalentity.Draft{GroupID: "group", WorkDate: date, Today: []journalentity.Work{}, Tomorrow: nil}, json: `{"GroupID":"group","WorkDate":"2026-11-23T00:00:00Z","Personnel":{"Management":null,"Control":null,"Workers":null},"Today":[],"Tomorrow":null}`, hash: "45f7115412e532c80fa0f6b29109fb9e167f20a190759e997dd7f5be9311368b", entry: `{"GroupID":"group","WorkDate":"2026-11-23T00:00:00Z","Personnel":{"Management":null,"Control":null,"Workers":null},"Today":[],"Tomorrow":null,"ID":"entry","CreatorID":"actor","CreatedAt":"2026-11-23T01:00:00Z"}`},
		{name: "repository", draft: journalentity.Draft{RepositoryID: "repo", WorkDate: date, Personnel: journalentity.Personnel{Management: []string{"manager"}, Control: []string{}, Workers: nil}, Today: []journalentity.Work{{Name: "工項", Percent: 45, Reference: &journalentity.Reference{Kind: "repository", ID: "repo"}, ObjectID: "object", SpaceIDs: []string{"space"}, ObjectName: "父 › 子", SpaceNames: []string{"空間"}}}, Tomorrow: []journalentity.Work{}}, json: `{"GroupID":"","RepositoryID":"repo","WorkDate":"2026-11-23T00:00:00Z","Personnel":{"Management":["manager"],"Control":[],"Workers":null},"Today":[{"Name":"工項","Percent":45,"Reference":{"Kind":"repository","ID":"repo"},"ObjectID":"object","SpaceIDs":["space"],"ObjectName":"父 › 子","SpaceNames":["空間"]}],"Tomorrow":[]}`, hash: "fa0ab71cbf8c84a8f1eec5ab1652b3e365e1b9487f2cf2ae9b4d7654f1d0c701", entry: `{"GroupID":"","RepositoryID":"repo","WorkDate":"2026-11-23T00:00:00Z","Personnel":{"Management":["manager"],"Control":[],"Workers":null},"Today":[{"Name":"工項","Percent":45,"Reference":{"Kind":"repository","ID":"repo"},"ObjectID":"object","SpaceIDs":["space"],"ObjectName":"父 › 子","SpaceNames":["空間"]}],"Tomorrow":[],"ID":"entry","CreatorID":"actor","CreatedAt":"2026-11-23T01:00:00Z"}`},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			body, err := json.Marshal(draftData(tc.draft))
			if err != nil || string(body) != tc.json || firestorekey.Digest(body) != tc.hash {
				t.Fatalf("receipt format changed: %s (%v)", body, err)
			}
			entry := journalentity.Entry{Draft: tc.draft, ID: "entry", CreatorID: "actor", CreatedAt: date.Add(time.Hour)}
			payload, err := json.Marshal(entryData(entry))
			if err != nil || string(payload) != tc.entry {
				t.Fatalf("stored format changed: %s (%v)", payload, err)
			}
			var stored entryPayload
			if err := json.Unmarshal([]byte(tc.entry), &stored); err != nil {
				t.Fatal(err)
			}
			if got := stored.model(); !reflect.DeepEqual(got, entry) {
				t.Fatalf("old payload changed on read: %+v", got)
			}
		})
	}
}
