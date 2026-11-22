package journalhttp

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	journaldto "github.com/963sup/line-mini-app-bot/internal/business/journal/application/dto"
	journalentity "github.com/963sup/line-mini-app-bot/internal/business/journal/domain/entity"
	"github.com/963sup/line-mini-app-bot/internal/identity"
	"github.com/963sup/line-mini-app-bot/internal/transport/http/api/generated"
)

type testAuth struct{}

func (testAuth) Authenticate(w http.ResponseWriter, r *http.Request) (string, string, bool) {
	if r.Header.Get("Authorization") != "Bearer test" {
		w.WriteHeader(401)
		return "", "", false
	}
	return "actor", "", true
}

type fixture struct {
	saved  journalentity.Entry
	filter journaldto.Filter
	key    string
}

func (f *fixture) Create(_ context.Context, actor, key string, d journalentity.Draft) (journalentity.Entry, error) {
	f.key = key
	f.saved = journalentity.Entry{Draft: d, ID: strings.Repeat("a", 32), CreatorID: actor, CreatedAt: d.WorkDate}
	return f.saved, nil
}

func (f *fixture) List(_ context.Context, actor string, filter journaldto.Filter) (journaldto.Page, error) {
	f.filter = filter
	return journaldto.Page{Items: []journalentity.Entry{f.saved}}, nil
}

func (f *fixture) Get(_ context.Context, actor, id string) (journalentity.Entry, error) {
	if id != f.saved.ID || actor != f.saved.CreatorID {
		return journalentity.Entry{}, journalentity.ErrNotFound
	}
	return f.saved, nil
}

func TestJournalRoutes(t *testing.T) {
	const base = "/api/v1/journals"
	const group = "C11111111111111111111111111111111"
	const key = "sample-journal-001"
	const repo = "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"
	const draft = `{"repositoryId":"` + repo + `","workDate":"2026-11-18","personnel":{"management":[],"control":[],"workers":["甲"]},"todayItems":[{"name":"鋪設","completionPercent":40,"reference":{"type":"project","id":"aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"}},{"name":"整理","completionPercent":100}],"tomorrowItems":[]}`
	f := new(fixture)
	mux := http.NewServeMux()
	Register(mux, f, testAuth{})
	call := func(method, path, body string, auth bool, want int) string {
		t.Helper()
		r := httptest.NewRequest(method, path, strings.NewReader(body))
		if auth {
			r.Header.Set("Authorization", "Bearer test")
		}
		if method == "POST" {
			r.Header.Set("Content-Type", "application/json")
			r.Header.Set("Idempotency-Key", key)
		}
		w := httptest.NewRecorder()
		mux.ServeHTTP(w, r)
		if w.Code != want {
			t.Fatalf("%s %s: %d %s", method, path, w.Code, w.Body.String())
		}
		return w.Body.String()
	}
	call("GET", "/api/v1/groups/"+group+"/journal-reference-targets", "", true, 404)
	call("GET", base, "", false, 401)
	call("POST", base, strings.Replace(draft, "100", "101", 1), true, 400)
	call("POST", base, strings.Replace(draft, `"repositoryId":"`+repo+`"`, `"groupId":"`+group+`"`, 1), true, 400)
	unselected := call("POST", base, draft, true, 201)
	for _, field := range []string{"objectId", "spaceIds", "objectName", "spaceNames"} {
		if strings.Contains(unselected, `"`+field+`"`) {
			t.Fatalf("unselected optional field %s must remain omitted: %s", field, unselected)
		}
	}
	if f.key != key || f.saved.CreatorID != "actor" || f.saved.RepositoryID != repo || f.saved.Today[0].Reference.Kind != "project" || f.saved.Today[1].Reference != nil {
		t.Fatal("creation mapping")
	}
	list := call("GET", base+"?workDate=2026-11-18&groupId="+group, "", true, 200)
	if f.filter.GroupID != group || f.filter.WorkDate != "2026-11-18" || !strings.Contains(list, f.saved.ID) {
		t.Fatal("dashboard filter or response")
	}
	call("GET", base+"/"+f.saved.ID, "", true, 200)
	call("GET", base+"/"+strings.Repeat("b", 32), "", true, 404)
	t.Run("work selection and server snapshots", func(t *testing.T) {
		const objectID = "bbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb"
		const spaceID = "dddddddddddddddddddddddddddddddd"
		selected := strings.Replace(draft, `"type":"project"`, `"type":"repository"`, 1)
		selected = strings.Replace(selected, `"completionPercent":40`, `"completionPercent":40,"objectId":"`+objectID+`","spaceIds":["`+spaceID+`"]`, 1)
		selected = strings.Replace(selected, `"tomorrowItems":[]`, `"tomorrowItems":[{"name":"明日","expectedCompletionPercent":50,"reference":{"type":"repository","id":"aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"},"objectId":"`+objectID+`","spaceIds":["`+spaceID+`"]}]`, 1)
		response := call("POST", base, selected, true, 201)
		if row := f.saved.Today[0]; row.ObjectID != objectID || len(row.SpaceIDs) != 1 || row.SpaceIDs[0] != spaceID {
			t.Fatalf("today selection lost at adapter: %+v", row)
		}
		if row := f.saved.Tomorrow[0]; row.ObjectID != objectID || len(row.SpaceIDs) != 1 || row.SpaceIDs[0] != spaceID {
			t.Fatalf("tomorrow selection lost at adapter: %+v", row)
		}
		if strings.Contains(list, `"progress"`) {
			t.Fatal("ordinary feed must omit progress")
		}
		chart := call("GET", base+"?repositoryId="+repo+"&includeProgress=true", "", true, 200)
		var dashboard generated.JournalPage
		if err := json.Unmarshal([]byte(chart), &dashboard); err != nil || len(dashboard.Items) != 1 || dashboard.Items[0].Progress == nil {
			t.Fatalf("missing dashboard progress: %s", chart)
		}
		rows := *dashboard.Items[0].Progress
		if len(rows) != 3 || rows[0].Name != "鋪設" || rows[0].Today == nil || *rows[0].Today != 40 || rows[0].Tomorrow != nil || rows[2].Tomorrow == nil || *rows[2].Tomorrow != 50 || rows[2].Today != nil {
			t.Fatalf("incorrect progress: %s", chart)
		}
		var created generated.Journal
		if err := json.Unmarshal([]byte(response), &created); err != nil || len(created.TodayItems) != 2 || len(created.TomorrowItems) != 1 || created.TodayItems[0].ObjectId != objectID || len(created.TodayItems[0].SpaceIds) != 1 || created.TomorrowItems[0].ObjectId != objectID {
			t.Fatalf("selection lost in response: %s (%v)", response, err)
		}
		f.saved.Today[0].ObjectName, f.saved.Today[0].SpaceNames = "父 › 子", []string{"作業區"}
		f.saved.Tomorrow[0].ObjectName, f.saved.Tomorrow[0].SpaceNames = "父", []string{"作業區"}
		response = call("GET", base+"/"+f.saved.ID, "", true, 200)
		var returned generated.Journal
		if err := json.Unmarshal([]byte(response), &returned); err != nil || len(returned.TodayItems) != 2 || len(returned.TomorrowItems) != 1 || returned.TodayItems[0].ObjectName != "父 › 子" || len(returned.TodayItems[0].SpaceNames) != 1 || returned.TodayItems[0].SpaceNames[0] != "作業區" || returned.TomorrowItems[0].ObjectName != "父" {
			t.Fatalf("server snapshots lost in response: %s (%v)", response, err)
		}
		spoofed := strings.Replace(selected, `"completionPercent":40`, `"completionPercent":40,"objectName":"偽造名稱","spaceNames":["偽造空間"]`, 1)
		call("POST", base, spoofed, true, 400)
		// The adapter itself must not promote client names to authoritative snapshots,
		// even when invoked without the request validator.
		r := httptest.NewRequest("POST", base, strings.NewReader(spoofed))
		r = r.WithContext(identity.WithUser(r.Context(), identity.User{ID: "actor"}))
		w := httptest.NewRecorder()
		(&journalServer{f}).CreateJournal(w, r, generated.CreateJournalParams{IdempotencyKey: key})
		if w.Code != 201 || f.saved.Today[0].ObjectName != "" || len(f.saved.Today[0].SpaceNames) != 0 {
			t.Fatalf("adapter accepted client snapshots: %d %+v", w.Code, f.saved.Today[0])
		}
	})
	mux = http.NewServeMux()
	Register(mux, nil, testAuth{})
	call("POST", base, draft, true, 503)
}
