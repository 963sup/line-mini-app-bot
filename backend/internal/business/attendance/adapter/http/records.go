package attendancehttp

import (
	"context"
	"errors"
	"net/http"
	"time"

	attendancedto "github.com/963sup/line-mini-app-bot/internal/business/attendance/application/dto"
	attendanceport "github.com/963sup/line-mini-app-bot/internal/business/attendance/application/port"
	attendancequery "github.com/963sup/line-mini-app-bot/internal/business/attendance/application/query"
	attendanceentity "github.com/963sup/line-mini-app-bot/internal/business/attendance/domain/entity"
	attendanceservice "github.com/963sup/line-mini-app-bot/internal/business/attendance/domain/service"
	transport "github.com/963sup/line-mini-app-bot/internal/transport/http"
	"github.com/963sup/line-mini-app-bot/internal/transport/http/api/generated"
)

// List and correction-detail responses omit the persisted location and event evidence.
type attendanceRecordView = generated.AttendanceRecord

func recordView(r attendanceentity.Record) attendanceRecordView {
	breaks := intervalViews(r.Breaks)
	work, status := attendanceservice.CalculateRecordAttendance(r)
	return generated.AttendanceRecord{AttendanceMinutes: &work.AttendanceMinutes, EstimatedMinutes: &work.EstimatedTotalMinutes, WorktimeStatus: &status, RecordID: r.RecordID, GroupID: r.GroupID, ClockInAt: r.ClockInAt, ClockOutAt: r.ClockOutAt, Revision: r.Revision, EventSequence: r.EventSequence, Breaks: &breaks, BreakStartedAt: r.BreakStartedAt, RestConfirmed: r.RestConfirmed, RebuildIssues: &r.RebuildIssues}
}

func (s *attendanceServer) query(w http.ResponseWriter, r *http.Request, serve func(context.Context, attendanceport.RecordQueries, string)) {
	bot, app := s.bot, s.app
	if bot == nil {
		transport.WriteAPIError(w, 503, "not_configured", "出勤尚未啟用。")
		return
	}
	user, _, ok := app.Authenticate(w, r)
	if !ok {
		return
	}
	store, ok := bot.store.(attendanceport.RecordQueries)
	if !ok {
		transport.WriteAPIError(w, 503, "not_configured", "出勤查詢尚未啟用。")
		return
	}
	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()
	serve(ctx, store, user)
}

func (s *attendanceServer) ListAttendanceGroupPage(w http.ResponseWriter, r *http.Request, params generated.ListAttendanceGroupPageParams) {
	s.query(w, r, func(ctx context.Context, store attendanceport.RecordQueries, user string) {
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
	s.query(w, r, func(ctx context.Context, store attendanceport.RecordQueries, user string) {
		group, month, ok := attendanceMonthInput(w, optionalString(params.Group), params.Month, s.bot, true)
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
	s.query(w, r, func(ctx context.Context, store attendanceport.RecordQueries, user string) {
		group, month, ok := attendanceMonthInput(w, params.Group, params.Month, s.bot, false)
		if !ok {
			return
		}
		report, err := attendancequery.ReadMonthlyReport(ctx, store, user, group, month)
		if attendanceQueryError(w, err) {
			return
		}
		calculation := calculationView(report.Calculation)
		calculation.OtherNoticeCount = pointerInt64(report.Calculation.OtherNoticeCount())
		transport.WriteJSON(w, 200, generated.MonthlySummary{Calculation: calculation, Count: int64(len(report.Records)), AsOf: time.Now().UTC()})
	})
}

func (s *attendanceServer) GetAttendanceRecord(w http.ResponseWriter, r *http.Request, id string, params generated.GetAttendanceRecordParams) {
	s.query(w, r, func(ctx context.Context, store attendanceport.RecordQueries, user string) {
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

func attendanceMonthInput(w http.ResponseWriter, group, value string, bot *Attendance, allowAllGroups bool) (string, time.Time, bool) {
	zone := bot.zone
	if zone == nil {
		zone = time.FixedZone("Asia/Taipei", 28800)
	}
	month, err := time.ParseInLocation("2006-01", value, zone)
	if err != nil || (!validAttendanceGroup(group) && (!allowAllGroups || group != "")) {
		transport.WriteAPIError(w, 400, "invalid_input", "群組或月份格式無效。")
		return "", time.Time{}, false
	}
	return group, month, true
}

func attendanceQueryError(w http.ResponseWriter, err error) bool {
	if err == nil {
		return false
	}
	switch {
	case errors.Is(err, attendancedto.ErrInvalidCursor):
		transport.WriteServiceError(w, 400, "invalid_cursor", "分頁已失效，請重新查詢。", err)
	case errors.Is(err, attendanceentity.ErrRecordNotFound):
		transport.WriteServiceError(w, 404, "not_found", "找不到此班次。", err)
	default:
		transport.WriteServiceError(w, 503, "unavailable", "暫時無法讀取出勤，請重試。", err)
	}
	return true
}

func (s *attendanceServer) GetAttendanceDailyHeadcount(w http.ResponseWriter, r *http.Request, params generated.GetAttendanceDailyHeadcountParams) {
	if s.bot == nil {
		transport.WriteAPIError(w, 503, "not_configured", "出勤尚未啟用。")
		return
	}
	user, _, ok := s.app.Authenticate(w, r)
	if !ok {
		return
	}
	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()
	if s.bot.membership == nil {
		transport.WriteAPIError(w, 503, "unavailable", "暫時無法查詢出勤參考。")
		return
	}
	member, err := s.bot.membership.IsMember(ctx, params.Group, user)
	if attendanceQueryError(w, err) {
		return
	}
	if !member {
		transport.WriteAPIError(w, 403, "not_member", "無法確認你是此群組成員。")
		return
	}
	reader, ok := s.bot.store.(attendanceport.DailyHeadcountReader)
	if !ok {
		transport.WriteAPIError(w, 503, "not_configured", "出勤人數查詢尚未啟用。")
		return
	}
	zone := time.FixedZone("Asia/Taipei", 8*60*60)
	year, month, day := params.Date.Date()
	start := time.Date(year, month, day, 0, 0, 0, 0, zone)
	count, err := reader.DailyHeadcount(ctx, params.Group, start, start.AddDate(0, 0, 1))
	if !attendanceQueryError(w, err) {
		transport.WriteJSON(w, 200, generated.AttendanceDailyHeadcount{Count: count, AsOf: time.Now().UTC()})
	}
}
