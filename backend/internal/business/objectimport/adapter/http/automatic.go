package objectimporthttp

import (
	"encoding/base64"
	application "github.com/963sup/line-mini-app-bot/internal/business/objectimport/application/automatic"
	"strings"

	workflow "github.com/963sup/line-mini-app-bot/internal/business/workflow/application"

	objectimportDomainEntity "github.com/963sup/line-mini-app-bot/internal/business/objectimport/domain/entity"
	objectimportDomainEntityAutomatic "github.com/963sup/line-mini-app-bot/internal/business/objectimport/domain/entity/automatic"
)

import (
	"errors"
	"net/http"

	"github.com/963sup/line-mini-app-bot/internal/identity"

	transport "github.com/963sup/line-mini-app-bot/internal/transport/http"
	"github.com/963sup/line-mini-app-bot/internal/transport/http/api/generated"
)

func automaticView(j objectimportDomainEntityAutomatic.Run) generated.ObjectImportRun {
	rows := make([]generated.ObjectImportRunRow, 0, len(j.Rows))
	var created, needsReview int
	for _, r := range j.Rows {
		if r.ResultObjectID != "" {
			created++
		}
		if r.Type == "needs_review" {
			needsReview++
		}
		rows = append(rows, generated.ObjectImportRunRow{Id: r.ID, Name: r.Name, Description: r.Description, Type: r.Type, Quantity: r.Quantity, Unit: r.Unit, MaterialCode: r.MaterialCode, SourceItemNo: r.SourceItemNo, SourcePages: r.SourcePages, ResultObjectId: r.ResultObjectID})
	}
	return generated.ObjectImportRun{CreatedCount: created, NeedsReviewCount: needsReview, CanReplaceSource: j.Status == "pending" && j.Reason != "" && len(j.Rows) == 0, Id: j.ID, RepositoryId: j.RepositoryID, CreatorId: j.CreatorID, Status: generated.ObjectImportRunStatus(j.Status), Reason: generated.ObjectImportRunReason(j.Reason), SourceDocument: j.SourceDocument, SourceRevision: j.SourceRevision, SourceHash: j.SourceHash, Rows: rows, CreatedAt: j.CreatedAt, UpdatedAt: j.UpdatedAt}
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
	source := objectimportDomainEntity.Source{SourceDocument: input.SourceDocument, SourceRevision: input.SourceRevision}
	if input.Pages != nil {
		for _, p := range *input.Pages {
			source.Pages = append(source.Pages, objectimportDomainEntity.Page{Number: p.Number, Text: p.Text})
		}
	}
	request := workflow.ImportRequest{Source: source, BatchType: selectedType(input.Type)}
	if input.Classification != nil {
		if string(*input.Classification) != "per_item" || input.Type != nil {
			writeAutomaticError(w, objectimportDomainEntity.ErrInput)
			return
		}
		request.PerItem = true
	}
	if input.Decisions != nil {
		request.Decisions = *input.Decisions
	}
	if input.Resume != nil {
		if !request.PerItem || input.SourceRevision != "" || input.Pages != nil || input.Workbook != nil || input.PdfFile != nil || input.TextFile != nil {
			writeAutomaticError(w, objectimportDomainEntity.ErrInput)
			return
		}
		corrections := map[string]application.Correction{}
		if input.Corrections != nil {
			for id, fields := range *input.Corrections {
				corrections[id] = application.Correction{Name: fields.Name, Quantity: fields.Quantity, Unit: fields.Unit}
			}
		}
		j, err := s.automatic.Resume(r.Context(), identity.UserID(r.Context()), rid, p.IdempotencyKey, input.SourceDocument, request.Decisions, corrections)
		if !writeAutomaticError(w, err) {
			transport.WriteJSON(w, 200, automaticView(j))
		}
		return
	}
	if input.Corrections != nil {
		writeAutomaticError(w, objectimportDomainEntity.ErrInput)
		return
	}
	payload := input.Workbook
	if input.PdfFile != nil {
		if payload != nil || input.TextFile != nil || !strings.HasPrefix(input.SourceDocument, "pdf:") {
			writeAutomaticError(w, objectimportDomainEntity.ErrInput)
			return
		}
		payload = input.PdfFile
	} else if input.TextFile != nil {
		if payload != nil || strings.HasPrefix(input.SourceDocument, "pdf:") || strings.HasPrefix(input.SourceDocument, "xlsx:") {
			writeAutomaticError(w, objectimportDomainEntity.ErrInput)
			return
		}
		payload = input.TextFile
	} else if payload != nil && !strings.HasPrefix(input.SourceDocument, "xlsx:") {
		writeAutomaticError(w, objectimportDomainEntity.ErrInput)
		return
	}
	if payload != nil {
		data, err := base64.StdEncoding.DecodeString(*payload)
		if err != nil || len(data) == 0 || input.Pages != nil {
			writeAutomaticError(w, objectimportDomainEntity.ErrInput)
			return
		}
		request.TableFile = data
	}
	j, e := workflow.Import(r.Context(), s.automatic, identity.UserID(r.Context()), rid, p.IdempotencyKey, request)
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
	case errors.Is(e, objectimportDomainEntity.ErrNotFound):
		status, code, message = 404, "not_found", "找不到此空間或匯入作業。"
	case errors.Is(e, objectimportDomainEntityAutomatic.ErrTypeRequired):
		status, code, message = 400, "type_required", "請先選擇施作或非施作，再以新的匯入識別碼送出。"
	case errors.Is(e, objectimportDomainEntity.ErrInput):
		status, code, message = 400, "invalid_input", "來源內容或大小不符合匯入規則。"
	case errors.Is(e, objectimportDomainEntity.ErrCreationConflict):
		status, code, message = 409, "idempotency_conflict", "識別碼已用於不同來源內容，請沿用原內容重試。"
	case errors.Is(e, objectimportDomainEntity.ErrForbidden):
		status, code, message = 403, "forbidden", "只有建立者可以操作此匯入作業。"
	case errors.Is(e, objectimportDomainEntity.ErrConflict):
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
