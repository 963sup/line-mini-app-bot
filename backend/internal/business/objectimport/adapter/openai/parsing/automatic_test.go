package openai

import (
	objectimportDomainEntity "github.com/963sup/line-mini-app-bot/internal/business/objectimport/domain/entity"
)

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"strings"
	"testing"
)

func TestAutomaticProviderNoQuoteSchemaAndPhysicalPages(t *testing.T) {
	const valid = `{"rows":[{"name":"Pump","quantity":"11","unit":"SET","sourceItemNo":"10","pages":[7]}]}`
	for _, content := range []string{valid, strings.Replace(valid, `"name":"Pump"`, `"type":"work","name":"Pump"`, 1), strings.Replace(valid, `"name":"Pump"`, `"description":"no","name":"Pump"`, 1), strings.Replace(valid, `"name":"Pump"`, `"materialCode":"no","name":"Pump"`, 1), strings.Replace(valid, `"pages":[7]`, `"pages":[7],"quote":"literal"`, 1), strings.Replace(valid, `"name":"Pump"`, `"actor":"administrator","name":"Pump"`, 1)} {
		p := Producer{Key: "fake", HTTP: &http.Client{Transport: transport(func(r *http.Request) (*http.Response, error) {
			b, _ := io.ReadAll(r.Body)
			var input map[string]any
			if json.Unmarshal(b, &input) != nil {
				t.Fatal("request")
			}
			messages := input["messages"].([]any)
			user := messages[1].(map[string]any)["content"].(string)
			if strings.Contains(user, "private-filename") || strings.Contains(user, "revision-private") {
				t.Fatal("metadata sent to provider")
			}
			prompt := messages[0].(map[string]any)["content"].(string)
			for _, forbidden := range []string{"Classify", "fee", "insurance", "profit", "non_work", "materialCode"} {
				if strings.Contains(prompt, forbidden) {
					t.Fatal("classification prompt remains", forbidden)
				}
			}
			schema := input["response_format"].(map[string]any)["json_schema"].(map[string]any)["schema"]
			properties := schema.(map[string]any)["properties"].(map[string]any)["rows"].(map[string]any)["items"].(map[string]any)["properties"].(map[string]any)
			if _, ok := properties["type"]; ok {
				t.Fatal("model selects type")
			}
			encoded, _ := json.Marshal(schema)
			if strings.Contains(string(encoded), "quote") || strings.Contains(string(encoded), `"description"`) || strings.Contains(string(encoded), `"materialCode"`) || strings.Contains(string(encoded), `"enum":["work"`) || !strings.Contains(string(encoded), `"enum":[7]`) {
				t.Fatal("wrong schema", string(encoded))
			}
			return response(content, "stop", "", nil), nil
		})}}
		rows, e := p.Rows(context.Background(), objectimportDomainEntity.Source{SourceDocument: "private-filename", SourceRevision: "revision-private", Pages: []objectimportDomainEntity.Page{{Number: 7, Text: "synthetic item"}}})
		if (content == valid) != (e == nil) {
			t.Fatal(rows, e)
		}
	}
}

func TestTextOnlyPDFUsesOneStructuredRequest(t *testing.T) {
	calls := 0
	p := Producer{Key: "fake", HTTP: &http.Client{Transport: transport(func(r *http.Request) (*http.Response, error) {
		calls++
		var request struct {
			Messages       []struct{ Content string }
			ResponseFormat json.RawMessage `json:"response_format"`
		}
		if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
			t.Fatal(err)
		}
		if len(request.Messages) != 2 || !strings.Contains(request.Messages[1].Content, "B-20 Lamp") ||
			!strings.Contains(string(request.ResponseFormat), `"strict":true`) ||
			!strings.Contains(string(request.ResponseFormat), `"enum":["needs_review"]`) {
			t.Fatal("expected one structured request using page text")
		}
		return response(`{"rows":[{"name":"Lamp","quantity":"","unit":"","sourceItemNo":"B-20","pages":[2],"type":"needs_review","quote":"B-20 Lamp"}]}`, "stop", "", nil), nil
	})}}
	rows, err := p.ClassifiedRows(t.Context(), objectimportDomainEntity.Source{PDF: objectimportDomainEntity.PDFPageText{}, Pages: []objectimportDomainEntity.Page{{Number: 2, Text: "B-20 Lamp"}}})
	if err != nil || calls != 1 || len(rows) != 1 || rows[0].Type != "needs_review" {
		t.Fatal(rows, calls, err)
	}
}
