package httpadapter

import (
	"context"
	"errors"
	"net/http"
	"strconv"
	"time"

	"github.com/963sup/line-mini-app-bot/internal/attendance"
	"github.com/963sup/line-mini-app-bot/internal/worktime"
)

type attendanceQueryStore interface {
	GroupsPage(context.Context, string, string, int) (attendance.GroupPage, error)
	RecordsPage(context.Context, string, string, time.Time, string, int) (attendance.RecordPage, error)
	Record(context.Context, string, string, string) (attendance.Record, error)
	MonthGroup(context.Context, string, string, time.Time) ([]attendance.Record, error)
}

// List and correction-detail responses omit the persisted location and event evidence.
type attendanceRecordView struct {
	RecordID       string
	GroupID        string
	ClockInAt      time.Time
	ClockOutAt     time.Time
	Revision       int64
	EventSequence  int64
	Breaks         []worktime.Interval
	BreakStartedAt time.Time
	RestConfirmed  bool
	RebuildIssues  []string
}

func recordView(r attendance.Record) attendanceRecordView {
	return attendanceRecordView{r.RecordID, r.GroupID, r.ClockInAt, r.ClockOutAt, r.Revision, r.EventSequence, r.Breaks, r.BreakStartedAt, r.RestConfirmed, r.RebuildIssues}
}

func registerAttendanceQueries(mux *http.ServeMux, bot *Attendance, app *miniApp) {
	register := func(pattern string, serve func(http.ResponseWriter, *http.Request, context.Context, attendanceQueryStore, string)) {
		mux.HandleFunc(pattern, func(w http.ResponseWriter, r *http.Request) {
			if bot == nil {
				writeAPIError(w, 503, "not_configured", "出勤尚未啟用。")
				return
			}
			user, _, ok := app.authenticate(w, r)
			if !ok {
				return
			}
			store, ok := bot.store.(attendanceQueryStore)
			if !ok {
				writeAPIError(w, 503, "not_configured", "出勤查詢尚未啟用。")
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
		type item struct {
			ID   string `json:"id"`
			Name string `json:"name"`
		}
		items := make([]item, 0, len(page.Items))
		for _, id := range page.Items {
			items = append(items, item{id, bot.historyGroupName(ctx, id)})
		}
		writeJSON(w, 200, struct {
			Items      []item    `json:"items"`
			NextCursor string    `json:"nextCursor"`
			AsOf       time.Time `json:"asOf"`
			QueryMode  string    `json:"queryMode"`
		}{items, page.NextCursor, time.Now().UTC(), page.QueryMode})
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
		writeJSON(w, 200, struct {
			Items      []attendanceRecordView `json:"items"`
			NextCursor string                 `json:"nextCursor"`
			AsOf       time.Time              `json:"asOf"`
			QueryMode  string                 `json:"queryMode"`
		}{items, page.NextCursor, time.Now().UTC(), page.QueryMode})
	})
	register("GET /api/v1/attendance/monthly/summary", func(w http.ResponseWriter, r *http.Request, ctx context.Context, store attendanceQueryStore, user string) {
		group, month, ok := attendanceMonthInput(w, r, bot)
		if !ok {
			return
		}
		records, err := store.MonthGroup(ctx, user, group, month)
		if attendanceQueryError(w, err) {
			return
		}
		report := attendance.CalculateMonthlyAttendance(records, month)
		writeJSON(w, 200, struct {
			Calculation worktime.Result
			Count       int       `json:"count"`
			AsOf        time.Time `json:"asOf"`
		}{report.Calculation, len(records), time.Now().UTC()})
	})
	register("GET /api/v1/attendance/records/{id}", func(w http.ResponseWriter, r *http.Request, ctx context.Context, store attendanceQueryStore, user string) {
		group := r.URL.Query().Get("group")
		id := r.PathValue("id")
		if !validAttendanceGroup(group) || !correctionIDPattern.MatchString(id) {
			writeAPIError(w, 400, "invalid_input", "請提供有效的群組與班次。")
			return
		}
		record, err := store.Record(ctx, user, group, id)
		if attendanceQueryError(w, err) {
			return
		}
		writeJSON(w, 200, recordView(record))
	})
}

func attendancePageInput(w http.ResponseWriter, r *http.Request) (int, string, bool) {
	size := 20
	if value := r.URL.Query().Get("pageSize"); value != "" {
		parsed, err := strconv.Atoi(value)
		if err != nil || parsed < 1 || parsed > 100 {
			writeAPIError(w, 400, "invalid_input", "每頁筆數須介於 1 至 100。")
			return 0, "", false
		}
		size = parsed
	}
	cursor := r.URL.Query().Get("cursor")
	if len(cursor) > 4096 {
		writeAPIError(w, 400, "invalid_cursor", "分頁已失效，請重新查詢。")
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
		writeAPIError(w, 400, "invalid_input", "請提供群組與 YYYY-MM 月份。")
		return "", time.Time{}, false
	}
	return group, month, true
}

func attendanceQueryError(w http.ResponseWriter, err error) bool {
	if err == nil {
		return false
	}
	switch {
	case errors.Is(err, attendance.ErrInvalidCursor):
		writeAPIError(w, 400, "invalid_cursor", "分頁已失效，請重新查詢。")
	case errors.Is(err, attendance.ErrRecordNotFound):
		writeAPIError(w, 404, "not_found", "找不到此班次。")
	default:
		writeAPIError(w, 503, "unavailable", "暫時無法讀取出勤，請重試。")
	}
	return true
}
