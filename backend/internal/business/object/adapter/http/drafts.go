package objecthttp

import (
	"context"
	"errors"
	"net/http"
	"time"

	object "github.com/963sup/line-mini-app-bot/internal/business/object/application"
	operations "github.com/963sup/line-mini-app-bot/internal/business/operations/application"
	"github.com/963sup/line-mini-app-bot/internal/identity"
	transport "github.com/963sup/line-mini-app-bot/internal/transport/http"
	"github.com/963sup/line-mini-app-bot/internal/transport/http/api/generated"
)

func (s *objectServer) PrepareObjectTemplateDraft(w http.ResponseWriter, r *http.Request, _ generated.PrepareObjectTemplateDraftParams) {
	var in generated.ObjectTemplateDraftRequest
	if !transport.ReadJSON(w, r, &in, "草稿請求無效。") {
		return
	}
	var data []byte
	if in.Data != nil {
		data = *in.Data
	}
	ctx, cancel := context.WithTimeout(r.Context(), 90*time.Second)
	defer cancel()
	result, err := s.service.PrepareTemplateDraft(ctx, identity.UserID(ctx), object.DraftPreparation{Source: string(in.Source), Filename: value(in.Filename), Data: data, Name: value(in.Name), Conditions: value(in.Conditions)})
	if err != nil {
		switch {
		case errors.Is(err, operations.ErrInvalidDocument):
			transport.WriteAPIError(w, 400, "invalid_document", "文件格式或大小無效。")
		case errors.Is(err, operations.ErrProvider):
			transport.WriteAPIError(w, 503, "provider_error", "草稿產生失敗，請稍後再試。")
		case errors.Is(err, operations.ErrIncomplete):
			transport.WriteAPIError(w, 503, "incomplete", "回應不完整，請縮小範圍後再試。")
		case errors.Is(err, operations.ErrRefused):
			transport.WriteAPIError(w, 400, "refused", "模型拒絕此請求。")
		case errors.Is(err, operations.ErrInvalidOutput):
			transport.WriteAPIError(w, 400, "invalid_output", "沒有可用的草稿，請補充條件或更換文件。")
		default:
			writeTemplateError(w, err)
		}
		return
	}
	nodes := make([]generated.ObjectTemplateDraftNode, len(result.Draft.Nodes))
	for i, n := range result.Draft.Nodes {
		nodes[i] = generated.ObjectTemplateDraftNode{Id: n.ID, Name: n.Name, ParentId: optional(n.ParentID), Unit: optional(n.Unit)}
	}
	candidates := make([]generated.ObjectCandidate, len(result.Candidates))
	for i, c := range result.Candidates {
		candidates[i] = generated.ObjectCandidate{Name: c.Name, Quote: c.Quote, Quantity: c.Quantity, Unit: c.Unit, Location: c.Location}
		if len(c.CategoryPath) > 0 {
			candidates[i].CategoryPath = &c.CategoryPath
		}
	}
	transport.WriteJSON(w, 200, generated.ObjectTemplateDraftResult{Draft: generated.ObjectTemplateDraft{Name: result.Draft.Name, Nodes: nodes}, Warnings: result.Warnings, Candidates: candidates})
}
