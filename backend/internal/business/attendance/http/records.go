package attendancehttp

import (
	"context"
	"errors"
	"net/http"
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

func (s *attendanceServer) query(w http.ResponseWriter, r *http.Request, serve func(context.Context, attendanceQueryStore, string)) {
	bot, app := s.bot, s.app
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
	serve(ctx, store, user)
}

func (s *attendanceServer) ListAttendanceGroupPage(w http.ResponseWriter, r *http.Request, params generated.ListAttendanceGroupPageParams) {
	s.query(w, r, func(ctx context.Context, store attendanceQueryStore, user string) {
		size, cursor, ok := attendancePageInput(w, params.PageSize, params.Cursor)
		if !ok {
			return
		}
		page, err := store.GroupsPage(ctx, user, cursor, size)
		if attendanceQueryError(w, err) {
			return
		}
		items := make([]generated.GroupItem, 0, len(page.Items))
		for _, id := range page.Items {
			items = append(items, generated.GroupItem{Id: id, Name: s.bot.historyGroupName(ctx, id)})
		}
		transport.WriteJSON(w, 200, generated.GroupPage{Items: items, NextCursor: page.NextCursor, AsOf: time.Now().UTC(), QueryMode: page.QueryMode})
	})
}

func (s *attendanceServer) ListAttendanceRecords(w http.ResponseWriter, r *http.Request, params generated.ListAttendanceRecordsParams) {
	s.query(w, r, func(ctx context.Context, store attendanceQueryStore, user string) {
		group, month, ok := attendanceMonthInput(w, params.Group, params.Month, s.bot)
		if !ok {
			return
		}
		size, cursor, ok := attendancePageInput(w, params.PageSize, params.Cursor)
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
}

func (s *attendanceServer) GetAttendanceMonthlySummary(w http.ResponseWriter, r *http.Request, params generated.GetAttendanceMonthlySummaryParams) {
	s.query(w, r, func(ctx context.Context, store attendanceQueryStore, user string) {
		group, month, ok := attendanceMonthInput(w, params.Group, params.Month, s.bot)
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
}

func (s *attendanceServer) GetAttendanceRecord(w http.ResponseWriter, r *http.Request, id string, params generated.GetAttendanceRecordParams) {
	s.query(w, r, func(ctx context.Context, store attendanceQueryStore, user string) {
		group := params.Group

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

func attendancePageInput(w http.ResponseWriter, pageSize *int, value *string) (int, string, bool) {
	size := 20
	if pageSize != nil {
		size = *pageSize
	}
	if size < 1 || size > 100 {
		transport.WriteAPIError(w, 400, "invalid_input", "每頁筆數須介於 1 至 100。")
		return 0, "", false
	}
	cursor := optionalString(value)
	if len(cursor) > 4096 {
		transport.WriteAPIError(w, 400, "invalid_cursor", "分頁已失效，請重新查詢。")
		return 0, "", false
	}
	return size, cursor, true
}

func optionalString(value *string) string {
	if value == nil {
		return ""
	}
	return *value
}

func validAttendanceGroup(group string) bool { return group != "" && len(group) <= 128 }

func attendanceMonthInput(w http.ResponseWriter, group, value string, bot *Attendance) (string, time.Time, bool) {
	zone := bot.zone
	if zone == nil {
		zone = time.FixedZone("Asia/Taipei", 28800)
	}
	month, err := time.ParseInLocation("2006-01", value, zone)
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
