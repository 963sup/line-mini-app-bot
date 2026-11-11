package attendancehttp

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"time"

	"github.com/963sup/line-mini-app-bot/internal/business/attendance"
	"github.com/963sup/line-mini-app-bot/internal/business/attendance/leave"
	"github.com/963sup/line-mini-app-bot/internal/httpapi/generated"
	transport "github.com/963sup/line-mini-app-bot/internal/transport/http"
)

func leaveView(l leave.Leave) generated.LeaveRecord {
	return generated.LeaveRecord{Id: l.ID, GroupId: l.GroupID, Kind: generated.LeaveRecordKind(l.Kind), Start: l.Start, End: l.End, Reason: l.Reason, CreatedAt: l.CreatedAt, Status: generated.Registered}
}

func registerLeave(mux *http.ServeMux, bot *Attendance, app Authenticator) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if bot == nil {
			transport.WriteAPIError(w, 503, "not_configured", "請假尚未啟用。")
			return
		}
		store, ok := bot.store.(leave.LeaveStore)
		if !ok {
			transport.WriteAPIError(w, 503, "not_configured", "請假尚未啟用。")
			return
		}
		user, _, ok := app.Authenticate(w, r)
		if !ok {
			return
		}
		ctx, cancel := context.WithTimeout(r.Context(), 10*time.Second)
		defer cancel()
		if r.Method == http.MethodGet {
			records, err := store.RecentLeaves(ctx, user)
			if err != nil {
				transport.WriteAPIError(w, 503, "unavailable", "暫時無法讀取請假紀錄。")
				return
			}
			items := make([]generated.LeaveRecord, 0, len(records))
			for _, l := range records {
				items = append(items, leaveView(l))
			}
			transport.WriteJSON(w, 200, items)
			return
		}
		r.Body = http.MaxBytesReader(w, r.Body, 4096)
		decoder := json.NewDecoder(r.Body)
		decoder.DisallowUnknownFields()
		var input generated.LeaveSubmission
		if decoder.Decode(&input) != nil || decoder.Decode(new(any)) != io.EOF || !launchTokenPattern.MatchString(input.Launch) || !leave.ValidLeaveKey(input.RequestKey) {
			transport.WriteAPIError(w, 400, "invalid_input", "請假資料無效。")
			return
		}
		proposal, err := leave.NormalizeLeave(leave.Leave{Kind: string(input.Kind), Start: input.Start, End: input.End, Reason: input.Reason})
		if err != nil {
			transport.WriteAPIError(w, 400, "invalid_input", "請確認假別、起訖時間及500字以內的理由；結束須晚於開始，期間最多366天。")
			return
		}
		group, err := store.LeaveGroup(ctx, input.Launch)
		if errors.Is(err, attendance.ErrExpiredLaunch) {
			transport.WriteAPIError(w, 403, "expired_launch", "請假連結已過期；請先查詢本人紀錄，或回原群組輸入「請假」。")
			return
		}
		if err != nil {
			transport.WriteAPIError(w, 503, "unavailable", "暫時無法確認請假連結。")
			return
		}
		if !requireGroupMember(w, ctx, group, user) {
			return
		}
		proposal.GroupID = group
		result, err := store.SubmitLeave(ctx, input.Launch, user, input.RequestKey, proposal)
		switch {
		case errors.Is(err, leave.ErrLeaveConflict):
			transport.WriteAPIError(w, 409, "idempotency_conflict", "同一重試識別碼已使用於不同內容，請先查詢原申請。")
		case errors.Is(err, attendance.ErrExpiredLaunch):
			transport.WriteAPIError(w, 403, "expired_launch", "請假連結已過期，請回原群組重新操作。")
		case errors.Is(err, leave.ErrInvalidLeave):
			transport.WriteAPIError(w, 400, "invalid_input", "請假內容無效。")
		case err != nil:
			transport.WriteAPIError(w, 503, "unavailable", "請假結果尚未確認；請查詢本人紀錄，重試時沿用原內容。")
		default:
			transport.WriteJSON(w, 200, leaveView(result))
		}
	})
	mux.Handle("GET /api/v1/attendance/leaves", handler)
	mux.Handle("POST /api/v1/attendance/leaves", http.NewCrossOriginProtection().Handler(handler))
}
