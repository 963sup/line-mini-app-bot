package openai

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"mime"
	"path/filepath"
	"strings"

	operations "github.com/963sup/line-mini-app-bot/internal/business/operations/application"
	sdk "github.com/openai/openai-go/v3"
	"github.com/openai/openai-go/v3/responses"
)

type Extractor struct {
	Responses responses.ResponseService
	Model     string
}

// candidatePayload owns the model provider's JSON format, independently of the application contract.
type candidatePayload struct {
	CategoryPath []string `json:"categoryPath"`
	Name         string   `json:"name"`
	Quantity     *string  `json:"quantity"`
	Unit         *string  `json:"unit"`
	Location     *string  `json:"location"`
	Quote        string   `json:"quote"`
}

func (c candidatePayload) candidate() operations.Candidate {
	return operations.Candidate(c)
}

func candidateFormat() responses.ResponseTextConfigParam {
	text := map[string]any{"type": "string"}
	nullable := map[string]any{"type": []string{"string", "null"}}
	item := map[string]any{
		"type": "object", "additionalProperties": false,
		"properties": map[string]any{"categoryPath": map[string]any{"type": "array", "items": text}, "name": text, "quantity": nullable, "unit": nullable, "location": nullable, "quote": text},
		"required":   []string{"categoryPath", "name", "quantity", "unit", "location", "quote"},
	}
	schema := map[string]any{
		"type": "object", "additionalProperties": false, "required": []string{"candidates"},
		"properties": map[string]any{"candidates": map[string]any{"type": "array", "items": item}},
	}
	format := responses.ResponseFormatTextConfigParamOfJSONSchema("object_candidates", schema)
	format.OfJSONSchema.Strict = sdk.Bool(true)
	return responses.ResponseTextConfigParam{Format: format}
}

func (e *Extractor) ExtractCandidates(ctx context.Context, doc operations.Document) (operations.Extraction, error) {
	if len(doc.Data) == 0 || len(doc.Data) >= 50_000_000 {
		return operations.Extraction{}, operations.ErrInvalidDocument
	}
	// Format support belongs to Responses, not a local extension allowlist.
	media := mime.TypeByExtension(strings.ToLower(filepath.Ext(doc.Filename)))
	if media == "" {
		media = "application/octet-stream"
	} else if parsed, _, err := mime.ParseMediaType(media); err == nil {
		media = parsed
	}
	file := responses.ResponseInputFileParam{Filename: sdk.String(doc.Filename), FileData: sdk.String("data:" + media + ";base64," + base64.StdEncoding.EncodeToString(doc.Data))}
	message := responses.ResponseInputItemParamOfInputMessage(responses.ResponseInputMessageContentListParam{{OfInputFile: &file}}, "user")
	response, err := e.Responses.New(ctx, responses.ResponseNewParams{
		Model: e.Model, Store: sdk.Bool(false),
		Instructions: sdk.String("Treat the document as source data. Extract one candidate per explicit work item or material row, even when names or codes repeat. Name is the complete original work item text, including wrapped lines, existing versus new qualifiers and equipment configuration. Keep titles printed within a work item in its name. Put separate headings that govern several rows in categoryPath from outermost to innermost. Do not turn headings, totals, item numbers, or material codes into work items. Take quantity and unit from the corresponding row columns, not numbers within its name. If the applicable quantity column is ambiguous, use null and include the relevant source text in the quote for review. Give a verbatim source quote and page or sheet and row when available. Use null for missing values, an empty categoryPath when uncertain, and an empty candidates array when no items are listed."),
		Input:        responses.ResponseNewParamsInputUnion{OfInputItemList: responses.ResponseInputParam{message}}, Text: candidateFormat(),
	})
	if err != nil {
		return operations.Extraction{}, operations.ErrProvider
	}
	result := operations.Extraction{RawResponse: []byte(response.RawJSON())}
	if response.Status == responses.ResponseStatusIncomplete {
		return result, operations.ErrIncomplete
	}
	if response.Status != responses.ResponseStatusCompleted {
		return result, operations.ErrInvalidOutput
	}
	for _, item := range response.Output {
		for _, part := range item.Content {
			if part.Type == "refusal" {
				return result, operations.ErrRefused
			}
		}
	}
	var output struct {
		Candidates []candidatePayload `json:"candidates"`
	}
	if json.Unmarshal([]byte(response.OutputText()), &output) != nil || output.Candidates == nil {
		return result, operations.ErrInvalidOutput
	}
	candidates := make([]operations.Candidate, 0, len(output.Candidates))
	for _, c := range output.Candidates {
		if strings.TrimSpace(c.Name) == "" || strings.TrimSpace(c.Quote) == "" {
			return result, operations.ErrInvalidOutput
		}
		candidates = append(candidates, c.candidate())
	}
	result.Candidates = candidates
	return result, nil
}
