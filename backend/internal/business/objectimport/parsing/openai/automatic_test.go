package openai

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"strings"
	"testing"

	feature "github.com/963sup/line-mini-app-bot/internal/business/objectimport"
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
		rows, e := p.Rows(context.Background(), feature.Source{SourceDocument: "private-filename", SourceRevision: "revision-private", Pages: []feature.Page{{Number: 7, Text: "synthetic item"}}})
		if (content == valid) != (e == nil) {
			t.Fatal(rows, e)
		}
	}
}
