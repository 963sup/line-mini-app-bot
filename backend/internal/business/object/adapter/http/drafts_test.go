package objecthttp

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	object "github.com/963sup/line-mini-app-bot/internal/business/object/application"
	operations "github.com/963sup/line-mini-app-bot/internal/business/operations/application"
)

type prepareTemplates struct {
	calls int
	size  int
}

func (p *prepareTemplates) PrepareTemplate(_ context.Context, actor string, in object.DraftPreparation) ([]operations.Candidate, error) {
	if actor != "A" {
		panic("missing trusted actor")
	}
	p.calls++
	p.size = len(in.Data)
	return []operations.Candidate{{Name: "基面處理", Quote: "推導建議"}}, nil
}

func TestPrepareTemplateHTTP(t *testing.T) {
	p := &prepareTemplates{}
	writes := &store{}
	invoke := func(body string, denied bool) *httptest.ResponseRecorder {
		mux := http.NewServeMux()
		Register(mux, &object.Service{TemplatePreparer: p, Store: writes}, auth{deny: denied})
		request := httptest.NewRequest("POST", "/api/v1/object-template-drafts", strings.NewReader(body))
		request.Header.Set("Content-Type", "application/json")
		response := httptest.NewRecorder()
		mux.ServeHTTP(response, request)
		return response
	}
	for _, tc := range []struct {
		body   string
		denied bool
		status int
	}{
		{`{"source":"procedure","name":"防水"}`, true, 401},
		{`{"source":"procedure"}`, false, 400},
		{`{"source":"procedure","name":"防水","actor":"other"}`, false, 400},
		{`{"source":"procedure","name":"防水"}`, false, 200},
	} {
		rr := invoke(tc.body, tc.denied)
		if rr.Code != tc.status || rr.Header().Get("Cache-Control") != "no-store" {
			t.Fatal(rr.Code, rr.Body.String())
		}
	}
	if p.calls != 1 || writes.calls != 0 {
		t.Fatal("preparation must not write", p.calls, writes.calls)
	}
	data := make([]byte, 70000)
	body, err := json.Marshal(map[string]any{"source": "document", "filename": "test.txt", "data": data})
	if err != nil {
		t.Fatal(err)
	}
	rr := invoke(string(body), false)
	if rr.Code != 200 || p.size != len(data) {
		t.Fatal(rr.Code, rr.Body.String(), p.size)
	}
}
