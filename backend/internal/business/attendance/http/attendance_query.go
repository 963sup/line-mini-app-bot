package attendancehttp

import (
	"context"
	"errors"
	"net/http"
	"strconv"
	"time"

	recordmodel "github.com/963sup/line-mini-app-bot/internal/business/attendance/records"
	"github.com/963sup/line-mini-app-bot/internal/httpapi/generated"
	transport "github.com/963sup/line-mini-app-bot/internal/transport/http"
)

type attendanceQueryStore interface {
	GroupsPage(ctx context.Context, userID, cursor string, pageSize int) (recordmodel.GroupPage, error)
	RecordsPage(ctx context.Context, userID, groupID string, month time.Time, cursor string, pageSize int) (recordmodel.RecordPage, error)
	Record(ctx context.Context, userID, groupID, recordID string) (recordmodel.Record, error)
	RecordsOverlappingMonth(ctx context.Context, userID, groupID string, month time.Time) ([]recordmodel.Record, error)
}

// List and correction-detail responses omit the persisted location and event evidence.
type attendanceRecordView = generated.AttendanceRecord

func recordView(r recordmodel.Record) attendanceRecordView {
	breaks := intervalViews(r.Breaks)
	work := recordmodel.CalculateRecordAttendance(r)
	status := "confirmed"
	if work.Uncalculable > 0 {
		status = "invalid"
	} else if work.Unconfirmed > 0 {
		status = "estimated"
	}
	return generated.AttendanceRecord{AttendanceMinutes: &work.AttendanceMinutes, EstimatedMinutes: &work.EstimatedTotalMinutes, WorktimeStatus: &status, RecordID: r.RecordID, GroupID: r.GroupID, ClockInAt: r.ClockInAt, ClockOutAt: r.ClockOutAt, Revision: r.Revision, EventSequence: r.EventSequence, Breaks: &breaks, BreakStartedAt: r.BreakStartedAt, RestConfirmed: r.RestConfirmed, RebuildIssues: &r.RebuildIssues}
}

func registerAttendanceQueries(mux *http.ServeMux, bot *Attendance, app Authenticator) {
	register := func(pattern string, serve func(http.ResponseWriter, *http.Request, context.Context, attendanceQueryStore, string)) {
		mux.HandleFunc(pattern, func(w http.ResponseWriter, r *http.Request) {
			if bot == nil {
				transport.WriteAPIError(w, 503, "not_configured", "出勤尚未啟用。")
				return
			}
			user, _, ok := app.Authenticate(w, r)
			if !ok {
				return
			}
			store, ok := bot.store.(attendanceQueryStore)
			if !ok {
				transport.WriteAPIError(w, 503, "not_configured", "出勤查詢尚未啟用。")
				return
			}
			ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
			defer cancel()
			serve(w, r, ctx, store, user)
		})
	}
	register("GET /api/v1/attendance/groups/page", func(w http.ResponseWriter, r *http.Request, ctx context.Context, store attendanceQueryStore, user string) {
		size, cursor, ok := attendancePageInput(w, r)
		if !ok {
			return
		}
		page, err := store.GroupsPage(ctx, user, cursor, size)
		if attendanceQueryError(w, err) {
			return
		}
		items := make([]generated.GroupItem, 0, len(page.Items))
		for _, id := range page.Items {
			items = append(items, generated.GroupItem{Id: id, Name: bot.historyGroupName(ctx, id)})
		}
		transport.WriteJSON(w, 200, generated.GroupPage{Items: items, NextCursor: page.NextCursor, AsOf: time.Now().UTC(), QueryMode: page.QueryMode})
	})
	register("GET /api/v1/attendance/records", func(w http.ResponseWriter, r *http.Request, ctx context.Context, store attendanceQueryStore, user string) {
		group, month, ok := attendanceMonthInput(w, r, bot)
		if !ok {
			return
		}
		size, cursor, ok := attendancePageInput(w, r)
		if !ok {
			return
		}
		page, err := store.RecordsPage(ctx, user, group, month, cursor, size)
		if attendanceQueryError(w, err) {
			return
		}
		items := make([]attendanceRecordView, 0, len(page.Items))
		for _, record := range page.Items {
			items = append(items, recordView(record))
		}
		transport.WriteJSON(w, 200, generated.RecordPage{Items: items, NextCursor: page.NextCursor, AsOf: time.Now().UTC(), QueryMode: page.QueryMode})
	})
	register("GET /api/v1/attendance/monthly/summary", func(w http.ResponseWriter, r *http.Request, ctx context.Context, store attendanceQueryStore, user string) {
		group, month, ok := attendanceMonthInput(w, r, bot)
		if !ok {
			return
		}
		records, err := store.RecordsOverlappingMonth(ctx, user, group, month)
		if attendanceQueryError(w, err) {
			return
		}
		report := recordmodel.CalculateMonthlyAttendance(records, month)
		transport.WriteJSON(w, 200, generated.MonthlySummary{Calculation: calculationView(report.Calculation), Count: int64(len(records)), AsOf: time.Now().UTC()})
	})
	register("GET /api/v1/attendance/records/{id}", func(w http.ResponseWriter, r *http.Request, ctx context.Context, store attendanceQueryStore, user string) {
		group := r.URL.Query().Get("group")
		id := r.PathValue("id")
		if !validAttendanceGroup(group) || !correctionIDPattern.MatchString(id) {
			transport.WriteAPIError(w, 400, "invalid_input", "請提供有效的群組與班次。")
			return
		}
		record, err := store.Record(ctx, user, group, id)
		if attendanceQueryError(w, err) {
			return
		}
		transport.WriteJSON(w, 200, recordView(record))
	})
}

func attendancePageInput(w http.ResponseWriter, r *http.Request) (int, string, bool) {
	size := 20
	if value := r.URL.Query().Get("pageSize"); value != "" {
		parsed, err := strconv.Atoi(value)
		if err != nil || parsed < 1 || parsed > 100 {
			transport.WriteAPIError(w, 400, "invalid_input", "每頁筆數須介於 1 至 100。")
			return 0, "", false
		}
		size = parsed
	}
	cursor := r.URL.Query().Get("cursor")
	if len(cursor) > 4096 {
		transport.WriteAPIError(w, 400, "invalid_cursor", "分頁已失效，請重新查詢。")
		return 0, "", false
	}
	return size, cursor, true
}

func validAttendanceGroup(group string) bool { return group != "" && len(group) <= 128 }

func attendanceMonthInput(w http.ResponseWriter, r *http.Request, bot *Attendance) (string, time.Time, bool) {
	zone := bot.zone
	if zone == nil {
		zone = time.FixedZone("Asia/Taipei", 28800)
	}
	group := r.URL.Query().Get("group")
	month, err := time.ParseInLocation("2006-01", r.URL.Query().Get("month"), zone)
	if err != nil || !validAttendanceGroup(group) {
		transport.WriteAPIError(w, 400, "invalid_input", "請提供群組與 YYYY-MM 月份。")
		return "", time.Time{}, false
	}
	return group, month, true
}

func attendanceQueryError(w http.ResponseWriter, err error) bool {
	if err == nil {
		return false
	}
	switch {
	case errors.Is(err, recordmodel.ErrInvalidCursor):
		transport.WriteAPIError(w, 400, "invalid_cursor", "分頁已失效，請重新查詢。")
	case errors.Is(err, recordmodel.ErrRecordNotFound):
		transport.WriteAPIError(w, 404, "not_found", "找不到此班次。")
	default:
		transport.WriteAPIError(w, 503, "unavailable", "暫時無法讀取出勤，請重試。")
	}
	return true
}
