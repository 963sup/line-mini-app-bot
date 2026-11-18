package objectimporthttp

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/963sup/line-mini-app-bot/internal/business/object"
	feature "github.com/963sup/line-mini-app-bot/internal/business/objectimport"
	"github.com/963sup/line-mini-app-bot/internal/business/objectimport/parsing"
	"github.com/963sup/line-mini-app-bot/internal/business/repository"
	"github.com/963sup/line-mini-app-bot/internal/httpapi/generated"
)

type auth struct{ deny bool }

func (a auth) Authenticate(w http.ResponseWriter, _ *http.Request) (string, string, bool) {
	if a.deny {
		w.WriteHeader(401)
		return "", "", false
	}
	return "A", "Actor", true
}

type repos struct{ deny bool }

func (r repos) Get(_ context.Context, _, id string) (repository.Repository, error) {
	if r.deny {
		return repository.Repository{}, repository.ErrNotFound
	}
	return repository.Repository{ID: id, GroupID: "C" + strings.Repeat("b", 32), Status: "ACTIVE"}, nil
}

type memory struct {
	job          feature.Job
	key, payload string
	calls        int
}

func (m *memory) Create(_ context.Context, _ repository.Repository, _, key, payload string, j feature.Job) (feature.Job, error) {
	m.calls++
	if m.key == key {
		if m.payload != payload {
			return feature.Job{}, feature.ErrCreationConflict
		}
		return m.job, nil
	}
	m.key = key
	m.payload = payload
	m.job = j
	return j, nil
}

func (m *memory) Get(context.Context, repository.Repository, string, string) (feature.Job, error) {
	m.calls++
	return m.job, nil
}

func (m *memory) Review(_ context.Context, _ repository.Repository, actor, _ string, input feature.Review) (feature.Job, error) {
	j, e := feature.ApplyReview(m.job, actor, input, time.Now().UTC())
	if e == nil {
		m.job = j
	}
	return j, e
}

func (m *memory) BeginConfirm(_ context.Context, _ repository.Repository, actor, _ string, v int64) (feature.Job, error) {
	j, e := feature.Begin(m.job, actor, v, time.Now().UTC())
	if e == nil {
		m.job = j
	}
	return j, e
}

func (m *memory) RecordResult(_ context.Context, _ repository.Repository, actor, _, row, id string, v int64) (feature.Job, error) {
	j, e := feature.Record(m.job, actor, row, id, v, time.Now().UTC())
	if e == nil {
		m.job = j
	}
	return j, e
}

type objects struct{ calls int }

func (o *objects) Create(_ context.Context, _, _, _ string, _ object.Content) (object.Object, error) {
	o.calls++
	return object.Object{ID: strings.Repeat("e", 32)}, nil
}

const valid = `{"sourceDocument":"source","sourceRevision":"v0","pages":[{"number":2,"text":"literal evidence; private remainder"}],"rows":[{"name":"Line","description":"proposal","type":"manual","quantity":"11.00","unit":"SET","materialCode":"","sourceItemNo":"10","evidence":[{"page":2,"quote":"literal evidence"}]}]}`

func TestHTTPDraftReviewConfirmAndStrictBoundaries(t *testing.T) {
	rid := strings.Repeat("a", 32)
	path := "/api/v1/repositories/" + rid + "/object-imports"
	m := &memory{}
	o := &objects{}
	mux := http.NewServeMux()
	Register(mux, &feature.Service{Repositories: repos{}, Objects: o, Store: m}, auth{})
	request := func(method, path, body string, key bool) *httptest.ResponseRecorder {
		t.Helper()
		r := httptest.NewRequest(method, path, strings.NewReader(body))
		r.Header.Set("Content-Type", "application/json")
		if key {
			r.Header.Set("Idempotency-Key", strings.Repeat("k", 16))
		}
		w := httptest.NewRecorder()
		mux.ServeHTTP(w, r)
		return w
	}
	for _, body := range []string{strings.TrimSuffix(valid, "}") + `,"actor":"B"}`, valid + ` {}`, strings.Replace(valid, `"quote":"literal evidence"`, `"quote":"absent"`, 1)} {
		w := request("POST", path, body, true)
		if w.Code != 400 || m.calls != 0 || o.calls != 0 {
			t.Fatalf("bad candidate reached write: %d %s", w.Code, w.Body.String())
		}
	}
	if w := request("POST", path, valid, false); w.Code != 400 {
		t.Fatal("missing idem accepted", w.Code)
	}
	w := request("POST", path, valid, true)
	if w.Code != 200 {
		t.Fatal(w.Code, w.Body.String())
	}
	if o.calls != 0 {
		t.Fatal("draft created objects")
	}
	var j generated.ObjectImport
	if e := json.Unmarshal(w.Body.Bytes(), &j); e != nil {
		t.Fatal(e)
	}
	if j.Status != "draft" || j.ReviewedAt != nil || j.Rows[0].Quantity != "11" || strings.Contains(w.Body.String(), "private remainder") {
		t.Fatal("draft result leaks raw page or mismatches", w.Body.String())
	}
	replay := request("POST", path, valid, true)
	var same generated.ObjectImport
	_ = json.Unmarshal(replay.Body.Bytes(), &same)
	if replay.Code != 200 || same.Id != j.Id {
		t.Fatal("draft idem failed")
	}
	if w = request("POST", path+"/"+j.Id+"/confirm", `{"expectedVersion":1}`, false); w.Code != 409 || o.calls != 0 {
		t.Fatal("unreviewed confirmation accepted")
	}
	review := generated.ObjectImportReview{ExpectedVersion: 1, Rows: []generated.ObjectImportReviewedRow{{Id: j.Rows[0].Id, Name: "Human corrected", Description: "reviewed", Type: "manual", Quantity: "11", Unit: "SET", MaterialCode: ""}}}
	b, _ := json.Marshal(review)
	w = request("POST", path+"/"+j.Id+"/review", string(b), false)
	if w.Code != 200 {
		t.Fatal(w.Code, w.Body.String())
	}
	w = request("POST", path+"/"+j.Id+"/confirm", `{"expectedVersion":2}`, false)
	if w.Code != 200 || o.calls != 1 {
		t.Fatal("confirmation", w.Code, w.Body.String())
	}
	var done generated.ObjectImport
	_ = json.Unmarshal(w.Body.Bytes(), &done)
	if done.Status != "completed" || done.Rows[0].ResultObjectId == "" {
		t.Fatal("missing result", done)
	}
	if w = request("POST", path+"/"+j.Id+"/confirm", `{"expectedVersion":2}`, false); w.Code != 200 || o.calls != 1 {
		t.Fatal("completed replay called objects again")
	}
	if w = request("POST", path+"/"+j.Id+"/review", string(b), false); w.Code != 409 {
		t.Fatal("edited frozen draft")
	}
}

func TestHTTPRejectsIdentityAndRepositoryBeforeStore(t *testing.T) {
	for _, tc := range []struct {
		auth, repo bool
		want       int
	}{{true, false, 401}, {false, true, 404}} {
		m := &memory{}
		mux := http.NewServeMux()
		Register(mux, &feature.Service{Repositories: repos{deny: tc.repo}, Store: m}, auth{deny: tc.auth})
		r := httptest.NewRequest("GET", "/api/v1/repositories/"+strings.Repeat("a", 32)+"/object-imports/"+strings.Repeat("b", 32), nil)
		w := httptest.NewRecorder()
		mux.ServeHTTP(w, r)
		if w.Code != tc.want || m.calls != 0 {
			t.Fatal(w.Code, w.Body.String())
		}
	}
}

type parserProducer struct {
	calls int
	quote string
	t     *testing.T
}

func (p *parserProducer) Candidates(ctx context.Context, _ feature.Source) ([]feature.Candidate, error) {
	p.calls++
	deadline, ok := ctx.Deadline()
	if !ok || time.Until(deadline) < 80*time.Second {
		p.t.Fatal("parse inherited short route timeout")
	}
	return []feature.Candidate{{Fields: feature.Fields{Name: "Line", Description: "proposal", Type: "manual", Quantity: "11.00", Unit: "SET"}, SourceItemNo: "10", Evidence: []feature.Evidence{{Page: 2, Quote: p.quote}}}}, ctx.Err()
}

func TestHTTPParseProducesOnlyValidatedCandidates(t *testing.T) {
	const body = `{"sourceDocument":"source","sourceRevision":"v0","pages":[{"number":2,"text":"literal evidence; private remainder"}],"consent":true}`
	cases := []struct {
		name                                                  string
		denyAuth, denyRepo, noConsent, missingModel, badQuote bool
		want                                                  int
	}{{name: "success", want: 200}, {name: "no identity", denyAuth: true, want: 401}, {name: "no membership", denyRepo: true, want: 404}, {name: "no consent", noConsent: true, want: 400}, {name: "not configured", missingModel: true, want: 503}, {name: "unsafe proposal", badQuote: true, want: 503}}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			drafts := &memory{}
			writer := &objects{}
			p := &parserProducer{quote: "literal evidence", t: t}
			if tc.badQuote {
				p.quote = "invented"
			}
			var producer parsing.Producer = p
			if tc.missingModel {
				producer = nil
			}
			parser := &parsing.Service{Repositories: repos{deny: tc.denyRepo}, Producer: producer}
			mux := http.NewServeMux()
			Register(mux, &feature.Service{Repositories: repos{}, Store: drafts, Objects: writer}, auth{deny: tc.denyAuth}, parser)
			text := body
			if tc.noConsent {
				text = strings.Replace(text, `"consent":true`, `"consent":false`, 1)
			}
			r := httptest.NewRequest("POST", "/api/v1/repositories/"+strings.Repeat("a", 32)+"/object-imports/parse", strings.NewReader(text))
			r.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()
			mux.ServeHTTP(w, r)
			if w.Code != tc.want {
				t.Fatal(w.Code, w.Body.String())
			}
			if drafts.calls != 0 || writer.calls != 0 {
				t.Fatal("parse wrote a job/Object")
			}
			if (tc.denyAuth || tc.denyRepo || tc.noConsent || tc.missingModel) && p.calls != 0 {
				t.Fatal("unauthorized/unconsented model call")
			}
			if tc.name == "success" {
				var output generated.ObjectImportParsed
				if e := json.Unmarshal(w.Body.Bytes(), &output); e != nil || len(output.Rows) != 1 || output.Rows[0].Quantity != "11" || len(output.SourceHash) != 64 {
					t.Fatal(output, e)
				}
				if strings.Contains(w.Body.String(), "private remainder") {
					t.Fatal("response includes full source text")
				}
			}
			if tc.missingModel && !strings.Contains(w.Body.String(), "not_configured") {
				t.Fatal("missing safe configuration error")
			}
		})
	}
}
