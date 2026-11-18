package objectimporthttp

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	feature "github.com/963sup/line-mini-app-bot/internal/business/objectimport"
	"github.com/963sup/line-mini-app-bot/internal/business/objectimport/automatic"
	"github.com/963sup/line-mini-app-bot/internal/business/repository"
	"github.com/963sup/line-mini-app-bot/internal/httpapi/generated"
)

type automaticMemory struct{ runs map[string]automatic.Run }

func (m *automaticMemory) Update(_ context.Context, r repository.Repository, a, id string, apply func(*automatic.Run) (automatic.Run, error)) (automatic.Run, error) {
	var old *automatic.Run
	if j, ok := m.runs[id]; ok {
		if j.CreatorID != a || j.RepositoryID != r.ID {
			return automatic.Run{}, feature.ErrNotFound
		}
		old = &j
	}
	j, e := apply(old)
	if e == nil {
		m.runs[id] = j
	}
	return j, e
}

type automaticProducer struct{ calls int }

func (p *automaticProducer) Rows(context.Context, feature.Source) ([]automatic.Candidate, error) {
	p.calls++
	return []automatic.Candidate{{Name: "Pump", Quantity: "11", Unit: "SET", SourceItemNo: "10", Pages: []int{7}}}, nil
}

func TestAutomaticHTTPKeyReplayAndBoundaries(t *testing.T) {
	for _, deny := range []bool{false, true} {
		m := &automaticMemory{runs: map[string]automatic.Run{}}
		p := &automaticProducer{}
		o := &objects{}
		s := &automatic.Service{Repositories: repos{}, Objects: o, Store: m, Producer: p}
		mux := http.NewServeMux()
		RegisterAutomatic(mux, &feature.Service{}, auth{deny: deny}, s, nil)
		path := "/api/v1/repositories/" + strings.Repeat("a", 32) + "/object-import-runs"
		body := `{"type":"work","sourceDocument":"pdf:synthetic","sourceRevision":"","pages":[{"number":7,"text":"synthetic pump"}]}`
		request := func(method, path, body, key string) *httptest.ResponseRecorder {
			r := httptest.NewRequest(method, path, strings.NewReader(body))
			r.Header.Set("Content-Type", "application/json")
			if key != "" {
				r.Header.Set("Idempotency-Key", key)
			}
			w := httptest.NewRecorder()
			mux.ServeHTTP(w, r)
			return w
		}
		w := request("POST", path, body, "stable-auto-key-001")
		if deny {
			if w.Code != 401 || p.calls != 0 {
				t.Fatal(w.Code)
			}
			continue
		}
		if w.Code != 200 {
			t.Fatal(w.Code, w.Body.String())
		}
		var j generated.ObjectImportRun
		if json.Unmarshal(w.Body.Bytes(), &j) != nil || j.Status != "completed" || len(j.Rows) != 1 || j.Rows[0].ResultObjectId == "" {
			t.Fatal(w.Body.String())
		}
		if w = request("POST", path, body, "stable-auto-key-001"); w.Code != 200 || p.calls != 1 || o.calls != 1 {
			t.Fatal(w.Code, p.calls, o.calls)
		}
		if w = request("GET", path+"/"+j.Id, "", ""); w.Code != 200 || w.Header().Get("Cache-Control") != "no-store" {
			t.Fatal(w.Code)
		}
		for _, bad := range []string{strings.Replace(body, `"sourceRevision":""`, `"sourceRevision":"changed"`, 1)} {
			if w = request("POST", path, bad, "stable-auto-key-001"); w.Code != 409 {
				t.Fatal(w.Code)
			}
		}
		if w = request("POST", path, strings.Replace(body, `"type":"work",`, "", 1), "missing-selection-key"); w.Code != 400 || !strings.Contains(w.Body.String(), "type_required") || len(m.runs) != 1 || p.calls != 1 {
			t.Fatal("missing selection", w.Code, w.Body.String())
		}
		if w = request("POST", path, strings.Replace(body, `"type":"work"`, `"type":"non_work"`, 1), "stable-auto-key-001"); w.Code != 409 {
			t.Fatal("changed type", w.Code)
		}
		if w = request("POST", path, body, ""); w.Code != 400 {
			t.Fatal(w.Code)
		}
		if w = request("POST", path, strings.Replace(body, `"sourceRevision":""`, `"actor":"admin","sourceRevision":""`, 1), "other-auto-key-001"); w.Code != 400 {
			t.Fatal(w.Code)
		}
	}
}
