package attendancehttp

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"math"
	"net/http"
	"net/http/httptest"
	"regexp"
	"strings"
	"testing"
	"time"

	recordmodel "github.com/963sup/line-mini-app-bot/internal/business/attendance/records"
)

func TestLocationNumberInRange(t *testing.T) {
	for _, tc := range []struct {
		name  string
		value float64
		valid bool
	}{
		{"minimum", -90, true},
		{"maximum", 90, true},
		{"zero", 0, true},
		{"below", -91, false},
		{"above", 91, false},
		{"NaN", math.NaN(), false},
		{"positive infinity", math.Inf(1), false},
		{"negative infinity", math.Inf(-1), false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			if got := locationNumberInRange(tc.value, -90, 90); got != tc.valid {
				t.Fatalf("valid=%v want=%v", got, tc.valid)
			}
		})
	}
}

func TestRequestIDPattern(t *testing.T) {
	for _, tc := range []struct {
		name, id string
		valid    bool
	}{
		{"hex", strings.Repeat("0123456789abcdef", 4), true},
		{"empty", "", false},
		{"short", strings.Repeat("a", 63), false},
		{"long", strings.Repeat("a", 65), false},
		{"uppercase", strings.Repeat("A", 64), false},
		{"nonhex", strings.Repeat("g", 64), false},
		{"newline", strings.Repeat("a", 64) + "\n", false},
		{"unicode", strings.Repeat("ａ", 64), false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			if got := requestIDPattern.MatchString(tc.id); got != tc.valid {
				t.Fatalf("valid=%v want=%v", got, tc.valid)
			}
		})
	}
}

func BenchmarkRequestIDValidation(b *testing.B) {
	id := strings.Repeat("0123456789abcdef", 4)
	b.Run("compile_per_request", func(b *testing.B) {
		b.ReportAllocs()
		for b.Loop() {
			if !regexp.MustCompile(`^[0-9a-f]{64}$`).MatchString(id) {
				b.Fatal("valid ID rejected")
			}
		}
	})
	b.Run("compiled_once", func(b *testing.B) {
		b.ReportAllocs()
		for b.Loop() {
			if !requestIDPattern.MatchString(id) {
				b.Fatal("valid ID rejected")
			}
		}
	})
}

// queryStoreStub exercises HTTP contracts without contacting LINE or Firestore.
type queryStoreStub struct{ attendanceStub }

func (s *queryStoreStub) GroupsPage(_ context.Context, user, _ string, _ int) (recordmodel.GroupPage, error) {
	s.queriedUser = user
	return recordmodel.GroupPage{Items: []string{"group"}, QueryMode: "legacy"}, s.err
}

func (s *queryStoreStub) RecordsPage(_ context.Context, user, group string, _ time.Time, _ string, _ int) (recordmodel.RecordPage, error) {
	s.queriedUser = user
	s.queriedGroup = group
	return recordmodel.RecordPage{Items: s.records, QueryMode: "legacy"}, s.err
}

func (s *queryStoreStub) Record(_ context.Context, user, group, _ string) (recordmodel.Record, error) {
	s.queriedUser = user
	s.queriedGroup = group
	if len(s.records) == 0 {
		return recordmodel.Record{}, recordmodel.ErrRecordNotFound
	}
	return s.records[0], s.err
}

func TestAttendanceQueryContracts(t *testing.T) {
	store := &queryStoreStub{attendanceStub: attendanceStub{records: []recordmodel.Record{{RecordID: "record", GroupID: "group", ClockInAt: time.Date(2026, 11, 1, 0, 0, 0, 0, time.UTC), ClockInPosition: &recordmodel.Position{Latitude: 25, Longitude: 121}}}}}
	app := testAuthenticator(&http.Client{Transport: memberTransport(func(r *http.Request) (*http.Response, error) {
		return &http.Response{StatusCode: 200, Body: io.NopCloser(strings.NewReader(fmt.Sprintf(`{"iss":"https://access.line.me","sub":"actor","aud":"123","exp":%d}`, time.Now().Add(time.Hour).Unix())))}, nil
	})})
	mux := http.NewServeMux()
	registerAttendanceQueries(mux, &Attendance{store: store}, app)
	for _, tc := range []struct {
		path   string
		status int
		auth   bool
		err    error
	}{
		{"/api/v1/attendance/groups/page", 401, false, nil},
		{"/api/v1/attendance/groups/page?pageSize=101", 400, true, nil},
		{"/api/v1/attendance/records?group=group&month=bad", 400, true, nil},
		{"/api/v1/attendance/records?group=group&month=2026-11&cursor=invalid", 400, true, recordmodel.ErrInvalidCursor},
		{"/api/v1/attendance/records?group=group&month=2026-11", 200, true, nil},
		{"/api/v1/attendance/records/record?group=group", 200, true, nil},
		{"/api/v1/attendance/monthly/summary?group=group&month=2026-11", 200, true, nil},
	} {
		t.Run(tc.path, func(t *testing.T) {
			store.err = tc.err
			request := httptest.NewRequest("GET", tc.path, nil)
			if tc.auth {
				request.Header.Set("Authorization", "Bearer token")
			}
			w := httptest.NewRecorder()
			mux.ServeHTTP(w, request)
			if w.Code != tc.status {
				t.Fatalf("status %d want %d: %s", w.Code, tc.status, w.Body.String())
			}
			if !json.Valid(w.Body.Bytes()) {
				t.Fatal("invalid JSON")
			}
			if tc.status == 200 {
				if store.queriedUser != "actor" {
					t.Fatal("query did not use verified actor")
				}
				if strings.Contains(w.Body.String(), "ClockInPosition") || strings.Contains(w.Body.String(), "Latitude") {
					t.Fatal("query exposed precise position")
				}
				if strings.Contains(tc.path, "/summary") && strings.Contains(w.Body.String(), `"Records"`) {
					t.Fatal("summary included raw records")
				}
			}
		})
	}
}
