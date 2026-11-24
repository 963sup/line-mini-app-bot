package operationshttp

import (
	"context"
	"encoding/json"
	"errors"
	"net/http/httptest"
	"strings"
	"testing"

	operations "github.com/963sup/line-mini-app-bot/internal/business/object/application/document"
	"github.com/963sup/line-mini-app-bot/internal/platform/identity"
	"github.com/963sup/line-mini-app-bot/internal/transport/http/api/generated"
)

type failedExtraction struct{ err error }

func (f failedExtraction) ExtractText(context.Context, operations.Document) (string, error) {
	return "", f.err
}

func TestExtractionErrorResponses(t *testing.T) {
	for _, tc := range []struct {
		err    error
		status int
		code   string
	}{
		{errors.Join(operations.ErrProvider, operations.ErrExtractorUnavailable), 503, "tool_unavailable"},
		{errors.Join(operations.ErrInvalidDocument, operations.ErrTextLimit), 422, "output_limit"},
		{errors.Join(operations.ErrInvalidDocument, operations.ErrDocumentPermission), 422, "pdf_permission"},
		{errors.Join(operations.ErrProvider, operations.ErrTextOutput), 500, "output_failed"},
		{errors.Join(operations.ErrProvider, operations.ErrTextEncoding), 500, "invalid_encoding"},
	} {
		t.Run(tc.code, func(t *testing.T) {
			h := Handler{Service: &operations.DocumentTextService{Extractor: failedExtraction{tc.err}}}
			req := httptest.NewRequest("POST", "/", strings.NewReader(`{"filename":"sample.pdf","data":"JVBERi0="}`))
			req = req.WithContext(identity.WithUser(req.Context(), identity.User{ID: "actor"}))
			w := httptest.NewRecorder()
			h.ExtractPopplerText(w, req, generated.ExtractPopplerTextParams{})
			var result struct{ Error struct{ Code string } }
			if err := json.Unmarshal(w.Body.Bytes(), &result); err != nil || w.Code != tc.status || result.Error.Code != tc.code {
				t.Fatalf("response %d %s", w.Code, w.Body.String())
			}
		})
	}
}
