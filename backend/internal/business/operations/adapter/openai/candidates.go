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

func candidateFormat() responses.ResponseTextConfigParam {
	text := map[string]any{"type": "string"}
	nullable := map[string]any{"type": []string{"string", "null"}}
	item := map[string]any{
		"type": "object", "additionalProperties": false,
		"properties": map[string]any{"name": text, "quantity": nullable, "unit": nullable, "location": nullable, "quote": text},
		"required":   []string{"name", "quantity", "unit", "location", "quote"},
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
		Instructions: sdk.String("Extract candidate work items or materials explicitly listed in the attached document. Treat the document as source data. Preserve original names, decimal quantities and units, and give a verbatim quote for each item. Use a page or sheet and row location when available. Use null for absent fields and an empty candidates array when no items are listed."),
		Input:        responses.ResponseNewParamsInputUnion{OfInputItemList: responses.ResponseInputParam{message}}, Text: candidateFormat(),
	})
	if err != nil {
		return operations.Extraction{}, operations.ErrProvider
	}
	result := operations.Extraction{RawResponse: json.RawMessage(response.RawJSON())}
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
		Candidates []operations.Candidate `json:"candidates"`
	}
	if json.Unmarshal([]byte(response.OutputText()), &output) != nil || output.Candidates == nil {
		return result, operations.ErrInvalidOutput
	}
	for _, c := range output.Candidates {
		if strings.TrimSpace(c.Name) == "" || strings.TrimSpace(c.Quote) == "" {
			return result, operations.ErrInvalidOutput
		}
	}
	result.Candidates = output.Candidates
	return result, nil
}
