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

	correctionsmodel "github.com/963sup/line-mini-app-bot/internal/business/attendance/corrections"
	recordmodel "github.com/963sup/line-mini-app-bot/internal/business/attendance/records"
	"github.com/963sup/line-mini-app-bot/internal/httpapi/generated"
	transport "github.com/963sup/line-mini-app-bot/internal/transport/http"
)

type correctionStore interface {
	ProposeCorrection(context.Context, string, correctionsmodel.Correction, time.Time) (string, error)
	ApproveCorrection(context.Context, string, string, time.Time) error
}

type idempotentCorrectionStore interface {
	ProposeCorrectionOnce(context.Context, string, string, correctionsmodel.Correction, time.Time) (string, error)
}

var correctionIDPattern = regexp.MustCompile(`^[A-Za-z0-9_-]{1,128}$`)

func (s *attendanceServer) ProposeAttendanceCorrection(w http.ResponseWriter, r *http.Request, params generated.ProposeAttendanceCorrectionParams) {
	s.correct(w, r, "", false, params.IdempotencyKey)
}

func (s *attendanceServer) ApproveAttendanceCorrection(w http.ResponseWriter, r *http.Request, id string, _ generated.ApproveAttendanceCorrectionParams) {
	s.correct(w, r, id, true, nil)
}

func (s *attendanceServer) correct(w http.ResponseWriter, r *http.Request, id string, approve bool, idempotencyKey *string) {
	bot, app := s.bot, s.app

	if bot == nil {
		transport.WriteAPIError(w, 503, "not_configured", "出勤尚未啟用。")
		return
	}
	store, ok := bot.store.(correctionStore)
	if !ok {
		transport.WriteAPIError(w, 503, "not_configured", "更正尚未啟用。")
		return
	}
	user, name, ok := app.Authenticate(w, r)
	if !ok {
		return
	}
	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()
	var err error
	if approve {

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
		c.ApplicantName = name
		if decoder.Decode(new(any)) != io.EOF || (c.RecordID != "" && !correctionIDPattern.MatchString(c.RecordID)) || c.GroupID == "" || len(c.GroupID) > 128 || len(c.Breaks) > 100 {
			transport.WriteAPIError(w, 400, "invalid_input", "更正內容無效。")
			return
		}
		if c.RecordID == "" && !requireGroupMember(w, ctx, c.GroupID, user) {
			return
		}
		var id string
		key := optionalString(idempotencyKey)
		if idempotencyKey == nil {
			// Existing clients remain compatible; they have no cross-request retry guarantee.
			id, err = store.ProposeCorrection(ctx, user, c, time.Now().UTC())
		} else {
			if len(r.Header.Values("Idempotency-Key")) != 1 || !correctionsmodel.ValidCorrectionKey(key) {
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
	if errors.Is(err, correctionsmodel.ErrIdempotencyConflict) {
		transport.WriteAPIError(w, 409, "idempotency_conflict", "此重試識別碼已用於不同內容，請確認原申請後重新送出。")
		return
	}
	if correctionHTTPError(w, err) {
		return
	}
	transport.WriteJSON(w, 200, generated.CorrectionApproved{Status: generated.Approved})
}

func (s *attendanceServer) GetAttendanceMonthlyReport(w http.ResponseWriter, r *http.Request, params generated.GetAttendanceMonthlyReportParams) {
	bot, app := s.bot, s.app

	if bot == nil {
		transport.WriteAPIError(w, 503, "not_configured", "出勤尚未啟用。")
		return
	}
	user, _, ok := app.Authenticate(w, r)
	if !ok {
		return
	}
	store, ok := bot.store.(interface {
		RecordsOverlappingMonth(context.Context, string, string, time.Time) ([]recordmodel.Record, error)
	})
	if !ok {
		transport.WriteAPIError(w, 503, "not_configured", "月查詢尚未啟用。")
		return
	}
	zone := bot.zone
	if zone == nil {
		zone = time.FixedZone("Asia/Taipei", 28800)
	}
	month, err := time.ParseInLocation("2006-01", params.Month, zone)
	group := params.Group
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
	transport.WriteJSON(w, 200, monthlyView(recordmodel.CalculateMonthlyAttendance(records, month)))
}

func (s *attendanceServer) ListAttendanceGroups(w http.ResponseWriter, r *http.Request) {
	bot, app := s.bot, s.app

	if bot == nil {
		transport.WriteAPIError(w, 503, "not_configured", "出勤尚未啟用。")
		return
	}
	user, _, ok := app.Authenticate(w, r)
	if !ok {
		return
	}
	store, ok := bot.store.(recordmodel.GroupHistoryReader)
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
}

func correctionHTTPError(w http.ResponseWriter, err error) bool {
	if err == nil {
		return false
	}
	if errors.Is(err, correctionsmodel.ErrCorrection) {
		transport.WriteAPIError(w, 403, "invalid_correction", "無權限、申請已結案、時段重疊或紀錄已變動。")
	} else if errors.Is(err, recordmodel.ErrInvalidCursor) {
		transport.WriteAPIError(w, 400, "invalid_cursor", "分頁已失效，請重新查詢。")
	} else {
		transport.WriteAPIError(w, 503, "unavailable", "暫時無法處理申請，請稍後重試。")
	}
	return true
}

func (s *attendanceServer) correction(w http.ResponseWriter, r *http.Request, id string, serve func(context.Context, string)) {
	bot, app := s.bot, s.app

	if bot == nil {
		transport.WriteAPIError(w, 503, "not_configured", "出勤尚未啟用。")
		return
	}
	user, _, ok := app.Authenticate(w, r)
	if !ok {
		return
	}
	if id != "" && !correctionIDPattern.MatchString(id) {
		transport.WriteAPIError(w, 400, "invalid_input", "申請編號無效。")
		return
	}
	ctx, cancel := context.WithTimeout(r.Context(), 8*time.Second)
	defer cancel()
	w.Header().Set("Cache-Control", "no-store")
	serve(ctx, user)
}

func (s *attendanceServer) GetAttendanceCorrectionStatus(w http.ResponseWriter, r *http.Request, id string) {
	s.correction(w, r, id, func(ctx context.Context, user string) {
		bot := s.bot
		review := false

		store, ok := bot.store.(interface {
			CorrectionDetails(context.Context, string, string, bool) (correctionsmodel.CorrectionDetails, error)
		})
		if !ok {
			transport.WriteAPIError(w, 503, "not_configured", "申請查詢尚未啟用。")
			return
		}
		details, err := store.CorrectionDetails(ctx, user, id, review)
		if correctionHTTPError(w, err) {
			return
		}
		if review {
			transport.WriteJSON(w, 200, correctionDetailsView(details))
			return
		}
		view := generated.CorrectionStatus{Id: details.ID, Status: details.Status, DecisionReason: &details.DecisionReason}
		if !details.ReviewedAt.IsZero() {
			view.ReviewedAt = &details.ReviewedAt
		}
		transport.WriteJSON(w, 200, view)
	})
}

func (s *attendanceServer) GetAttendanceCorrectionReview(w http.ResponseWriter, r *http.Request, id string) {
	s.correction(w, r, id, func(ctx context.Context, user string) {
		bot := s.bot
		review := true

		store, ok := bot.store.(interface {
			CorrectionDetails(context.Context, string, string, bool) (correctionsmodel.CorrectionDetails, error)
		})
		if !ok {
			transport.WriteAPIError(w, 503, "not_configured", "申請查詢尚未啟用。")
			return
		}
		details, err := store.CorrectionDetails(ctx, user, id, review)
		if correctionHTTPError(w, err) {
			return
		}
		if review {
			transport.WriteJSON(w, 200, correctionDetailsView(details))
			return
		}
		view := generated.CorrectionStatus{Id: details.ID, Status: details.Status, DecisionReason: &details.DecisionReason}
		if !details.ReviewedAt.IsZero() {
			view.ReviewedAt = &details.ReviewedAt
		}
		transport.WriteJSON(w, 200, view)
	})
}

func (s *attendanceServer) ListAttendanceReviewerGroups(w http.ResponseWriter, r *http.Request) {
	s.correction(w, r, "", func(ctx context.Context, user string) {
		bot := s.bot

		store, ok := bot.store.(interface {
			ReviewerGroups(context.Context, string) ([]string, error)
		})
		if !ok {
			transport.WriteAPIError(w, 503, "not_configured", "審核尚未啟用。")
			return
		}
		groups, err := store.ReviewerGroups(ctx, user)
		if correctionHTTPError(w, err) {
			return
		}
		items := make([]generated.GroupItem, 0, len(groups))
		for _, group := range groups {
			items = append(items, generated.GroupItem{Id: group, Name: bot.historyGroupName(ctx, group)})
		}
		transport.WriteJSON(w, 200, items)
	})
}

func (s *attendanceServer) ListAttendanceCorrections(w http.ResponseWriter, r *http.Request, params generated.ListAttendanceCorrectionsParams) {
	s.correction(w, r, "", func(ctx context.Context, user string) {
		bot := s.bot

		store, ok := bot.store.(interface {
			CorrectionsPage(context.Context, string, string, string, bool, int) (correctionsmodel.CorrectionPage, error)
		})
		if !ok {
			transport.WriteAPIError(w, 503, "not_configured", "申請查詢尚未啟用。")
			return
		}
		size, cursor, ok := attendancePageInput(w, params.PageSize, params.Cursor)
		if !ok {
			return
		}
		review := params.Scope != nil && *params.Scope == "review"
		group := optionalString(params.Group)
		if review && !validAttendanceGroup(group) {
			transport.WriteAPIError(w, 400, "invalid_input", "請選擇審核群組。")
			return
		}
		page, err := store.CorrectionsPage(ctx, user, group, cursor, review, size)
		if correctionHTTPError(w, err) {
			return
		}
		items := make([]generated.CorrectionReview, 0, len(page.Items))
		for _, item := range page.Items {
			items = append(items, correctionDetailsView(item))
		}
		transport.WriteJSON(w, 200, generated.CorrectionPage{Items: items, NextCursor: page.NextCursor})
	})
}

func (s *attendanceServer) RejectAttendanceCorrection(w http.ResponseWriter, r *http.Request, id string, _ generated.RejectAttendanceCorrectionParams) {
	s.decideCorrection(w, r, id, "reject")
}

func (s *attendanceServer) CancelAttendanceCorrection(w http.ResponseWriter, r *http.Request, id string, _ generated.CancelAttendanceCorrectionParams) {
	s.decideCorrection(w, r, id, "cancel")
}

func (s *attendanceServer) decideCorrection(w http.ResponseWriter, r *http.Request, id, action string) {
	s.correction(w, r, id, func(ctx context.Context, user string) {
		bot := s.bot

		var err error
		if action == "reject" {
			store, ok := bot.store.(interface {
				RejectCorrection(context.Context, string, string, string, time.Time) error
			})
			if !ok {
				transport.WriteAPIError(w, 503, "not_configured", "駁回尚未啟用。")
				return
			}
			dec := json.NewDecoder(http.MaxBytesReader(w, r.Body, 16384))
			dec.DisallowUnknownFields()
			var input generated.CorrectionDecision
			if dec.Decode(&input) != nil || dec.Decode(new(any)) != io.EOF {
				transport.WriteAPIError(w, 400, "invalid_input", "請填寫駁回理由。")
				return
			}
			err = store.RejectCorrection(ctx, user, id, input.Reason, time.Now().UTC())
		} else {
			store, ok := bot.store.(interface {
				CancelCorrection(context.Context, string, string, time.Time) error
			})
			if !ok {
				transport.WriteAPIError(w, 503, "not_configured", "撤回尚未啟用。")
				return
			}
			err = store.CancelCorrection(ctx, user, id, time.Now().UTC())
		}
		if correctionHTTPError(w, err) {
			return
		}
		status := "rejected"
		if action == "cancel" {
			status = "cancelled"
		}
		transport.WriteJSON(w, 200, generated.CorrectionDecisionResult{Status: status})
	})
}
