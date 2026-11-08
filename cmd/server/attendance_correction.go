package main

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"mime"
	"net/http"
	"regexp"
	"time"

	"github.com/963sup/line-mini-app-bot/internal/attendance"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type correctionStore interface {
	ProposeCorrection(context.Context, string, attendance.Correction, time.Time) (string, error)
	ApproveCorrection(context.Context, string, string, time.Time) error
}

var correctionIDPattern = regexp.MustCompile(`^[A-Za-z0-9_-]{1,128}$`)

func registerAttendanceCorrections(mux *http.ServeMux, bot *attendanceBot, app *miniApp) {
	handler := func(approve bool) http.Handler {
		return http.NewCrossOriginProtection().Handler(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if bot == nil {
				writeAPIError(w, 503, "not_configured", "出勤尚未啟用。")
				return
			}
			store, ok := bot.store.(correctionStore)
			if !ok {
				writeAPIError(w, 503, "not_configured", "更正尚未啟用。")
				return
			}
			user, _, ok := app.authenticate(w, r)
			if !ok {
				return
			}
			ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
			defer cancel()
			var err error
			if approve {
				id := r.PathValue("id")
				if !correctionIDPattern.MatchString(id) {
					writeAPIError(w, 400, "invalid_input", "申請編號無效。")
					return
				}
				err = store.ApproveCorrection(ctx, user, id, time.Now().UTC())
			} else {
				media, _, e := mime.ParseMediaType(r.Header.Get("Content-Type"))
				if e != nil || media != "application/json" {
					writeAPIError(w, 415, "invalid_content_type", "請使用 JSON。")
					return
				}
				r.Body = http.MaxBytesReader(w, r.Body, 16384)
				decoder := json.NewDecoder(r.Body)
				decoder.DisallowUnknownFields()
				var c attendance.Correction
				if decoder.Decode(&c) != nil || decoder.Decode(new(any)) != io.EOF || !correctionIDPattern.MatchString(c.RecordID) || c.GroupID == "" || len(c.GroupID) > 128 || len(c.Breaks) > 100 {
					writeAPIError(w, 400, "invalid_input", "更正內容無效。")
					return
				}
				var id string
				id, err = store.ProposeCorrection(ctx, user, c, time.Now().UTC())
				if err == nil {
					writeJSON(w, 202, map[string]string{"id": id, "status": "pending"})
					return
				}
			}
			if errors.Is(err, attendance.ErrCorrection) || status.Code(err) == codes.NotFound {
				writeAPIError(w, 403, "invalid_correction", "無權限、資料版本已變動或更正內容無效。")
				return
			}
			if err != nil {
				writeAPIError(w, 503, "unavailable", "暫時無法處理更正，請稍後重試。")
				return
			}
			writeJSON(w, 200, map[string]string{"status": "approved"})
		}))
	}
	mux.HandleFunc("GET /api/v1/attendance/monthly", func(w http.ResponseWriter, r *http.Request) {
		if bot == nil {
			writeAPIError(w, 503, "not_configured", "出勤尚未啟用。")
			return
		}
		user, _, ok := app.authenticate(w, r)
		if !ok {
			return
		}
		store, ok := bot.store.(interface {
			MonthGroup(context.Context, string, string, time.Time) ([]attendance.Record, error)
		})
		if !ok {
			writeAPIError(w, 503, "not_configured", "月查詢尚未啟用。")
			return
		}
		zone := bot.zone
		if zone == nil {
			zone = time.FixedZone("Asia/Taipei", 28800)
		}
		month, err := time.ParseInLocation("2006-01", r.URL.Query().Get("month"), zone)
		group := r.URL.Query().Get("group")
		if err != nil || group == "" || len(group) > 128 {
			writeAPIError(w, 400, "invalid_input", "請提供群組與 YYYY-MM 月份。")
			return
		}
		ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
		defer cancel()
		records, err := store.MonthGroup(ctx, user, group, month)
		if err != nil {
			writeAPIError(w, 503, "unavailable", "暫時無法讀取紀錄。")
			return
		}
		w.Header().Set("Cache-Control", "no-store")
		writeJSON(w, 200, attendance.CalculateMonthlyAttendance(records, month))
	})

	mux.HandleFunc("GET /api/v1/attendance/corrections/{id}", func(w http.ResponseWriter, r *http.Request) {
		if bot == nil {
			writeAPIError(w, 503, "not_configured", "出勤尚未啟用。")
			return
		}
		user, _, ok := app.authenticate(w, r)
		if !ok {
			return
		}
		store, ok := bot.store.(interface {
			CorrectionStatus(context.Context, string, string) (string, error)
		})
		if !ok {
			writeAPIError(w, 503, "not_configured", "更正查詢尚未啟用。")
			return
		}
		id := r.PathValue("id")
		if !correctionIDPattern.MatchString(id) {
			writeAPIError(w, 400, "invalid_input", "申請編號無效。")
			return
		}
		ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
		defer cancel()
		state, err := store.CorrectionStatus(ctx, user, id)
		if errors.Is(err, attendance.ErrCorrection) || status.Code(err) == codes.NotFound {
			writeAPIError(w, 403, "invalid_correction", "無法查詢此申請。")
			return
		}
		if err != nil {
			writeAPIError(w, 503, "unavailable", "暫時無法查詢。")
			return
		}
		w.Header().Set("Cache-Control", "no-store")
		writeJSON(w, 200, map[string]string{"id": id, "status": state})
	})
	mux.HandleFunc("GET /api/v1/attendance/groups", func(w http.ResponseWriter, r *http.Request) {
		if bot == nil {
			writeAPIError(w, 503, "not_configured", "出勤尚未啟用。")
			return
		}
		user, _, ok := app.authenticate(w, r)
		if !ok {
			return
		}
		store, ok := bot.store.(attendance.LocatedStore)
		if !ok {
			writeAPIError(w, 503, "not_configured", "查詢尚未啟用。")
			return
		}
		ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
		defer cancel()
		groups, err := store.Groups(ctx, user)
		if err != nil {
			writeAPIError(w, 503, "unavailable", "暫時無法查詢群組。")
			return
		}
		items := []map[string]string{}
		for _, id := range groups {
			items = append(items, map[string]string{"id": id, "name": bot.historyGroupName(ctx, id)})
		}
		w.Header().Set("Cache-Control", "no-store")
		writeJSON(w, 200, items)
	})
	mux.HandleFunc("GET /api/v1/attendance/corrections/{id}/review", func(w http.ResponseWriter, r *http.Request) {
		if bot == nil {
			writeAPIError(w, 503, "not_configured", "出勤尚未啟用。")
			return
		}
		user, _, ok := app.authenticate(w, r)
		if !ok {
			return
		}
		store, ok := bot.store.(interface {
			CorrectionReview(context.Context, string, string) (attendance.Correction, error)
		})
		if !ok {
			writeAPIError(w, 503, "not_configured", "審核尚未啟用。")
			return
		}
		id := r.PathValue("id")
		if !correctionIDPattern.MatchString(id) {
			writeAPIError(w, 400, "invalid_input", "申請編號無效。")
			return
		}
		ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
		defer cancel()
		c, err := store.CorrectionReview(ctx, user, id)
		if errors.Is(err, attendance.ErrCorrection) || status.Code(err) == codes.NotFound {
			writeAPIError(w, 403, "forbidden", "無權審核此申請。")
			return
		}
		if err != nil {
			writeAPIError(w, 503, "unavailable", "暫時無法讀取申請。")
			return
		}
		w.Header().Set("Cache-Control", "no-store")
		writeJSON(w, 200, c)
	})

	mux.Handle("POST /api/v1/attendance/corrections", handler(false))
	mux.Handle("POST /api/v1/attendance/corrections/{id}/approve", handler(true))
}
