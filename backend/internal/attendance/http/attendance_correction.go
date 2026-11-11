package attendancehttp

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"mime"
	"net/http"
	"regexp"
	"time"

	transport "github.com/963sup/line-mini-app-bot/internal/transport/http"

	"github.com/963sup/line-mini-app-bot/internal/attendance"
	"github.com/963sup/line-mini-app-bot/internal/httpapi/generated"
)

type correctionStore interface {
	ProposeCorrection(context.Context, string, attendance.Correction, time.Time) (string, error)
	ApproveCorrection(context.Context, string, string, time.Time) error
}

type idempotentCorrectionStore interface {
	ProposeCorrectionOnce(context.Context, string, string, attendance.Correction, time.Time) (string, error)
}

var correctionIDPattern = regexp.MustCompile(`^[A-Za-z0-9_-]{1,128}$`)

func registerAttendanceCorrections(mux *http.ServeMux, bot *Attendance, app Authenticator) {
	handler := func(approve bool) http.Handler {
		return http.NewCrossOriginProtection().Handler(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if bot == nil {
				transport.WriteAPIError(w, 503, "not_configured", "出勤尚未啟用。")
				return
			}
			store, ok := bot.store.(correctionStore)
			if !ok {
				transport.WriteAPIError(w, 503, "not_configured", "更正尚未啟用。")
				return
			}
			user, _, ok := app.Authenticate(w, r)
			if !ok {
				return
			}
			ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
			defer cancel()
			var err error
			if approve {
				id := r.PathValue("id")
				if !correctionIDPattern.MatchString(id) {
					transport.WriteAPIError(w, 400, "invalid_input", "申請編號無效。")
					return
				}
				err = store.ApproveCorrection(ctx, user, id, time.Now().UTC())
			} else {
				media, _, e := mime.ParseMediaType(r.Header.Get("Content-Type"))
				if e != nil || media != "application/json" {
					transport.WriteAPIError(w, 415, "invalid_content_type", "請使用 JSON。")
					return
				}
				r.Body = http.MaxBytesReader(w, r.Body, 16384)
				decoder := json.NewDecoder(r.Body)
				decoder.DisallowUnknownFields()
				var body generated.CorrectionSubmission
				if decoder.Decode(&body) != nil {
					transport.WriteAPIError(w, 400, "invalid_input", "更正內容無效。")
					return
				}
				c := correctionInput(body)
				if decoder.Decode(new(any)) != io.EOF || !correctionIDPattern.MatchString(c.RecordID) || c.GroupID == "" || len(c.GroupID) > 128 || len(c.Breaks) > 100 {
					transport.WriteAPIError(w, 400, "invalid_input", "更正內容無效。")
					return
				}
				var id string
				key := r.Header.Get("Idempotency-Key")
				if len(r.Header.Values("Idempotency-Key")) == 0 {
					// Existing clients remain compatible; they have no cross-request retry guarantee.
					id, err = store.ProposeCorrection(ctx, user, c, time.Now().UTC())
				} else {
					if len(r.Header.Values("Idempotency-Key")) != 1 || !attendance.ValidCorrectionKey(key) {
						transport.WriteAPIError(w, 400, "invalid_idempotency_key", "重試識別碼無效。")
						return
					}
					once, ok := bot.store.(idempotentCorrectionStore)
					if !ok {
						transport.WriteAPIError(w, 503, "not_configured", "安全重試尚未啟用。")
						return
					}
					id, err = once.ProposeCorrectionOnce(ctx, user, key, c, time.Now().UTC())
				}
				if err == nil {
					transport.WriteJSON(w, 202, generated.CorrectionAccepted{Id: id, Status: generated.Accepted})
					return
				}
			}
			if errors.Is(err, attendance.ErrIdempotencyConflict) {
				transport.WriteAPIError(w, 409, "idempotency_conflict", "此重試識別碼已用於不同內容，請確認原申請後重新送出。")
				return
			}
			if errors.Is(err, attendance.ErrCorrection) {
				transport.WriteAPIError(w, 403, "invalid_correction", "無權限、資料版本已變動或更正內容無效。")
				return
			}
			if err != nil {
				transport.WriteAPIError(w, 503, "unavailable", "暫時無法處理更正，請稍後重試。")
				return
			}
			transport.WriteJSON(w, 200, generated.CorrectionApproved{Status: generated.Approved})
		}))
	}
	mux.HandleFunc("GET /api/v1/attendance/monthly", func(w http.ResponseWriter, r *http.Request) {
		if bot == nil {
			transport.WriteAPIError(w, 503, "not_configured", "出勤尚未啟用。")
			return
		}
		user, _, ok := app.Authenticate(w, r)
		if !ok {
			return
		}
		store, ok := bot.store.(interface {
			RecordsOverlappingMonth(context.Context, string, string, time.Time) ([]attendance.Record, error)
		})
		if !ok {
			transport.WriteAPIError(w, 503, "not_configured", "月查詢尚未啟用。")
			return
		}
		zone := bot.zone
		if zone == nil {
			zone = time.FixedZone("Asia/Taipei", 28800)
		}
		month, err := time.ParseInLocation("2006-01", r.URL.Query().Get("month"), zone)
		group := r.URL.Query().Get("group")
		if err != nil || group == "" || len(group) > 128 {
			transport.WriteAPIError(w, 400, "invalid_input", "請提供群組與 YYYY-MM 月份。")
			return
		}
		ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
		defer cancel()
		records, err := store.RecordsOverlappingMonth(ctx, user, group, month)
		if err != nil {
			transport.WriteAPIError(w, 503, "unavailable", "暫時無法讀取紀錄。")
			return
		}
		w.Header().Set("Cache-Control", "no-store")
		transport.WriteJSON(w, 200, monthlyView(attendance.CalculateMonthlyAttendance(records, month)))
	})

	mux.HandleFunc("GET /api/v1/attendance/corrections/{id}", func(w http.ResponseWriter, r *http.Request) {
		if bot == nil {
			transport.WriteAPIError(w, 503, "not_configured", "出勤尚未啟用。")
			return
		}
		user, _, ok := app.Authenticate(w, r)
		if !ok {
			return
		}
		store, ok := bot.store.(interface {
			CorrectionStatus(context.Context, string, string) (string, error)
		})
		if !ok {
			transport.WriteAPIError(w, 503, "not_configured", "更正查詢尚未啟用。")
			return
		}
		id := r.PathValue("id")
		if !correctionIDPattern.MatchString(id) {
			transport.WriteAPIError(w, 400, "invalid_input", "申請編號無效。")
			return
		}
		ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
		defer cancel()
		state, err := store.CorrectionStatus(ctx, user, id)
		if errors.Is(err, attendance.ErrCorrection) {
			transport.WriteAPIError(w, 403, "invalid_correction", "無法查詢此申請。")
			return
		}
		if err != nil {
			transport.WriteAPIError(w, 503, "unavailable", "暫時無法查詢。")
			return
		}
		w.Header().Set("Cache-Control", "no-store")
		transport.WriteJSON(w, 200, generated.CorrectionStatus{Id: id, Status: state})
	})
	mux.HandleFunc("GET /api/v1/attendance/groups", func(w http.ResponseWriter, r *http.Request) {
		if bot == nil {
			transport.WriteAPIError(w, 503, "not_configured", "出勤尚未啟用。")
			return
		}
		user, _, ok := app.Authenticate(w, r)
		if !ok {
			return
		}
		store, ok := bot.store.(attendance.GroupHistoryReader)
		if !ok {
			transport.WriteAPIError(w, 503, "not_configured", "查詢尚未啟用。")
			return
		}
		ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
		defer cancel()
		groups, err := store.Groups(ctx, user)
		if err != nil {
			transport.WriteAPIError(w, 503, "unavailable", "暫時無法查詢群組。")
			return
		}
		items := []generated.GroupItem{}
		for _, id := range groups {
			items = append(items, generated.GroupItem{Id: id, Name: bot.historyGroupName(ctx, id)})
		}
		w.Header().Set("Cache-Control", "no-store")
		transport.WriteJSON(w, 200, items)
	})
	mux.HandleFunc("GET /api/v1/attendance/corrections/{id}/review", func(w http.ResponseWriter, r *http.Request) {
		if bot == nil {
			transport.WriteAPIError(w, 503, "not_configured", "出勤尚未啟用。")
			return
		}
		user, _, ok := app.Authenticate(w, r)
		if !ok {
			return
		}
		store, ok := bot.store.(interface {
			CorrectionReview(context.Context, string, string) (attendance.Correction, error)
		})
		if !ok {
			transport.WriteAPIError(w, 503, "not_configured", "審核尚未啟用。")
			return
		}
		id := r.PathValue("id")
		if !correctionIDPattern.MatchString(id) {
			transport.WriteAPIError(w, 400, "invalid_input", "申請編號無效。")
			return
		}
		ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
		defer cancel()
		c, err := store.CorrectionReview(ctx, user, id)
		if errors.Is(err, attendance.ErrCorrection) {
			transport.WriteAPIError(w, 403, "forbidden", "無權審核此申請。")
			return
		}
		if err != nil {
			transport.WriteAPIError(w, 503, "unavailable", "暫時無法讀取申請。")
			return
		}
		w.Header().Set("Cache-Control", "no-store")
		transport.WriteJSON(w, 200, correctionView(c))
	})

	mux.Handle("POST /api/v1/attendance/corrections", handler(false))
	mux.Handle("POST /api/v1/attendance/corrections/{id}/approve", handler(true))
}
