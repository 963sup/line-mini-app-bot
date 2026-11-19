package objectimporthttp

import (
	"errors"
	"net/http"

	"github.com/963sup/line-mini-app-bot/internal/identity"

	feature "github.com/963sup/line-mini-app-bot/internal/business/objectimport"
	"github.com/963sup/line-mini-app-bot/internal/business/objectimport/automatic"
	"github.com/963sup/line-mini-app-bot/internal/httpapi/generated"
	transport "github.com/963sup/line-mini-app-bot/internal/transport/http"
)

func automaticView(j automatic.Run) generated.ObjectImportRun {
	rows := make([]generated.ObjectImportRunRow, 0, len(j.Rows))
	for _, r := range j.Rows {
		rows = append(rows, generated.ObjectImportRunRow{Id: r.ID, Name: r.Name, Description: r.Description, Type: r.Type, Quantity: r.Quantity, Unit: r.Unit, MaterialCode: r.MaterialCode, SourceItemNo: r.SourceItemNo, SourcePages: r.SourcePages, ResultObjectId: r.ResultObjectID})
	}
	return generated.ObjectImportRun{Id: j.ID, RepositoryId: j.RepositoryID, CreatorId: j.CreatorID, Status: generated.ObjectImportRunStatus(j.Status), Reason: generated.ObjectImportRunReason(j.Reason), SourceDocument: j.SourceDocument, SourceRevision: j.SourceRevision, SourceHash: j.SourceHash, Rows: rows, CreatedAt: j.CreatedAt, UpdatedAt: j.UpdatedAt}
}

func (s *importServer) CreateObjectImportRun(w http.ResponseWriter, r *http.Request, rid string, p generated.CreateObjectImportRunParams) {
	if s.automatic == nil {
		transport.WriteAPIError(w, 503, "not_configured", "物件匯入尚未啟用。")
		return
	}
	var input generated.ObjectImportRunCreation
	if !transport.ReadJSON(w, r, &input, "草稿 JSON 無效。") {
		return
	}
	source := feature.Source{SourceDocument: input.SourceDocument, SourceRevision: input.SourceRevision}
	for _, p := range input.Pages {
		source.Pages = append(source.Pages, feature.Page{Number: p.Number, Text: p.Text})
	}
	var j automatic.Run
	var e error
	if input.Classification != nil {
		if string(*input.Classification) != "per_item" || input.Type != nil {
			writeAutomaticError(w, feature.ErrInput)
			return
		}
		var decisions map[string]string
		if input.Decisions != nil {
			decisions = *input.Decisions
		}
		j, e = s.automatic.Classify(r.Context(), identity.UserID(r.Context()), rid, p.IdempotencyKey, source, decisions)
	} else if input.Decisions != nil {
		e = feature.ErrInput
	} else {
		j, e = s.automatic.Create(r.Context(), identity.UserID(r.Context()), rid, p.IdempotencyKey, source, selectedType(input.Type))
	}
	if !writeAutomaticError(w, e) {
		transport.WriteJSON(w, 200, automaticView(j))
	}
}

func (s *importServer) GetObjectImportRun(w http.ResponseWriter, r *http.Request, rid, id string) {
	if s.automatic == nil {
		transport.WriteAPIError(w, 503, "not_configured", "物件匯入尚未啟用。")
		return
	}
	j, e := s.automatic.Get(r.Context(), identity.UserID(r.Context()), rid, id)
	if !writeAutomaticError(w, e) {
		transport.WriteJSON(w, 200, automaticView(j))
	}
}

func writeAutomaticError(w http.ResponseWriter, e error) bool {
	if e == nil {
		return false
	}
	status, code, message := 503, "unavailable", "匯入結果尚未確認，請沿用原識別碼與原內容重試。"
	switch {
	case errors.Is(e, feature.ErrNotFound):
		status, code, message = 404, "not_found", "找不到此空間或匯入作業。"
	case errors.Is(e, automatic.ErrTypeRequired):
		status, code, message = 400, "type_required", "請先選擇施作或非施作，再以新的匯入識別碼送出。"
	case errors.Is(e, feature.ErrInput):
		status, code, message = 400, "invalid_input", "來源內容或大小不符合匯入規則。"
	case errors.Is(e, feature.ErrCreationConflict):
		status, code, message = 409, "idempotency_conflict", "識別碼已用於不同來源內容，請沿用原內容重試。"
	case errors.Is(e, feature.ErrForbidden):
		status, code, message = 403, "forbidden", "只有建立者可以操作此匯入作業。"
	case errors.Is(e, feature.ErrConflict):
		status, code, message = 409, "state_conflict", "匯入狀態不符，請重新讀取。"
	}
	transport.WriteServiceError(w, status, code, message, e)
	return true
}

func selectedType(value *generated.ObjectImportRunCreationType) string {
	if value == nil {
		return ""
	}
	return string(*value)
}
