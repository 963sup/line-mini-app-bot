package operationshttp

import (
	"context"
	"errors"
	"mime"
	"net/http"

	"github.com/963sup/line-mini-app-bot/internal/business/operations/adapter/poppler"
	operations "github.com/963sup/line-mini-app-bot/internal/business/operations/application"
	"github.com/963sup/line-mini-app-bot/internal/identity"
	transport "github.com/963sup/line-mini-app-bot/internal/transport/http"
	httpapi "github.com/963sup/line-mini-app-bot/internal/transport/http/api"
	"github.com/963sup/line-mini-app-bot/internal/transport/http/api/generated"
)

type (
	Authenticator interface {
		Authenticate(http.ResponseWriter, *http.Request) (string, string, bool)
	}
	Handler struct {
		Service *operations.DocumentTextService
	}
)

func Register(mux *http.ServeMux, service *operations.DocumentTextService, auth Authenticator) {
	httpapi.RegisterPoppler(mux, &Handler{Service: service}, func(next http.Handler) http.Handler {
		return http.NewCrossOriginProtection().Handler(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Cache-Control", "no-store")
			actor, name, ok := auth.Authenticate(w, r)
			if !ok {
				return
			}
			media, _, err := mime.ParseMediaType(r.Header.Get("Content-Type"))
			if err != nil || media != "application/json" {
				transport.WriteAPIError(w, 415, "invalid_content_type", "請使用 JSON。")
				return
			}
			r.Body = http.MaxBytesReader(w, r.Body, 12<<20)
			next.ServeHTTP(w, r.WithContext(identity.WithUser(r.Context(), identity.User{ID: actor, Name: name})))
		}))
	})
}

func (h *Handler) ExtractPopplerText(w http.ResponseWriter, r *http.Request, _ generated.ExtractPopplerTextParams) {
	var in generated.PopplerDocument
	if !transport.ReadJSON(w, r, &in, "PDF 資料無效或超過大小限制。") {
		return
	}
	if in.Filename == "" || len(in.Filename) > 255 || len(in.Data) == 0 || len(in.Data) > 8<<20 {
		transport.WriteAPIError(w, 400, "invalid_document", "請選擇 8 MiB 以內的非空 PDF。")
		return
	}
	text, err := h.Service.ExtractText(r.Context(), identity.UserID(r.Context()), operations.Document{Filename: in.Filename, Data: in.Data})
	if err == nil {
		transport.WriteJSON(w, 200, generated.PopplerText{Text: text})
		return
	}
	status, code, message := 500, "conversion_failed", "PDF 轉換失敗。"
	switch {
	case errors.Is(err, context.DeadlineExceeded):
		status, code, message = 504, "extraction_timeout", "PDF 轉換逾時，請縮小文件後再試。"
	case errors.Is(err, context.Canceled):
		return
	case errors.Is(err, poppler.ErrUnavailable):
		status, code, message = 503, "tool_unavailable", "PDF 文字擷取工具尚未安裝。"
	case errors.Is(err, poppler.ErrOutputLimit):
		status, code, message = 422, "output_limit", "擷取文字超過 8 MiB 上限。"
	case errors.Is(err, poppler.ErrPermission):
		status, code, message = 422, "pdf_permission", "PDF 權限不允許擷取文字。"
	case errors.Is(err, operations.ErrInvalidDocument):
		status, code, message = 400, "invalid_document", "PDF 無效、損壞或無法開啟。"
	case errors.Is(err, poppler.ErrOpenOutput):
		code, message = "output_failed", "無法讀取轉換輸出。"
	case errors.Is(err, poppler.ErrEncoding):
		code, message = "invalid_encoding", "轉換輸出不是有效 UTF-8。"
	}
	transport.WriteAPIError(w, status, code, message)
}
