package attendancehttp

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/963sup/line-mini-app-bot/internal/httpapi"

	"github.com/963sup/line-mini-app-bot/internal/attendance/worktime"

	"github.com/963sup/line-mini-app-bot/internal/attendance"
	"github.com/getkin/kin-openapi/openapi3filter"
	"github.com/getkin/kin-openapi/routers/gorillamux"
)

type contractStore struct {
	attendanceStub
	reads, writes int
	key           string
	conflict      bool
}

func (s *contractStore) GroupsPage(context.Context, string, string, int) (attendance.GroupPage, error) {
	s.reads++
	return attendance.GroupPage{}, nil
}

func (s *contractStore) RecordsPage(context.Context, string, string, time.Time, string, int) (attendance.RecordPage, error) {
	s.reads++
	return attendance.RecordPage{}, nil
}

func (s *contractStore) Record(context.Context, string, string, string) (attendance.Record, error) {
	s.reads++
	return attendance.Record{RecordID: "r1", GroupID: "g"}, nil
}

func (s *contractStore) ProposeCorrection(context.Context, string, attendance.Correction, time.Time) (string, error) {
	s.writes++
	return "c1", nil
}

func (s *contractStore) ProposeCorrectionOnce(_ context.Context, _ string, key string, _ attendance.Correction, _ time.Time) (string, error) {
	s.writes++
	s.key = key
	if s.conflict {
		return "", attendance.ErrIdempotencyConflict
	}
	return "c1", nil
}

func (s *contractStore) ApproveCorrection(context.Context, string, string, time.Time) error {
	s.writes++
	return nil
}

func (s *contractStore) CorrectionStatus(context.Context, string, string) (string, error) {
	s.reads++
	return "pending", nil
}

func (s *contractStore) CorrectionReview(context.Context, string, string) (attendance.Correction, error) {
	s.reads++
	return attendance.Correction{ID: "c1", RecordID: "r1", GroupID: "g"}, nil
}

func (s *contractStore) CreateLaunch(context.Context, string, string) (string, error) { return "", nil }

func (s *contractStore) LaunchGroup(context.Context, string) (string, error) {
	s.reads++
	return "", attendance.ErrExpiredLaunch
}

func (s *contractStore) ExchangeLaunch(context.Context, string, string) (string, error) {
	s.writes++
	return "", nil
}

func contractTestHandler(t *testing.T, store *contractStore) (http.Handler, *int) {
	t.Helper()
	authCalls := new(int)
	app := testAuthenticator(&http.Client{Transport: memberTransport(func(r *http.Request) (*http.Response, error) {
		*authCalls++
		if r.URL.Host != "api.line.me" {
			t.Fatal("unexpected identity host")
		}
		body := fmt.Sprintf(`{"iss":"https://access.line.me","sub":"u1","aud":"123","exp":%d,"name":"Member"}`, time.Now().Add(time.Hour).Unix())
		return &http.Response{StatusCode: 200, Body: io.NopCloser(strings.NewReader(body)), Header: make(http.Header)}, nil
	})})
	mux := http.NewServeMux()
	registerAttendanceAPIWithIdentity(mux, &Attendance{store: store, zone: time.UTC}, app)
	return mux, authCalls
}

func contractRequest(method, path, body string) *http.Request {
	req := httptest.NewRequest(method, path, strings.NewReader(body))
	req.Header.Set("Authorization", "Bearer private-token")
	if method == http.MethodPost {
		req.Header.Set("Content-Type", "application/json")
	}
	return req
}

func TestContractRejectsInvalidRequestsBeforeStore(t *testing.T) {
	for _, tc := range []struct{ method, path, body, code string }{
		{"GET", "/api/v1/attendance/records?group=g&month=2026-11&pageSize=101", "", "invalid_input"},
		{"GET", "/api/v1/attendance/records?group=g&month=2026-13", "", "invalid_input"},
		{"GET", "/api/v1/attendance/records?group=g&month=2026-11&cursor=" + strings.Repeat("a", 4097), "", "invalid_cursor"},
		{"POST", "/api/v1/attendance/launch", `{"launch":"secret-invalid"}`, "invalid_launch"},
		{"POST", "/api/v1/attendance/punches", `{"requestId":"` + strings.Repeat("a", 64) + `","latitude":91,"longitude":121,"accuracyMeters":1,"capturedAt":"2026-11-10T00:00:00Z"}`, "invalid_position"},
		{"POST", "/api/v1/attendance/corrections", `{"groupId":"g","recordId":"r1","private-extra":"sensitive"}`, "invalid_input"},
	} {
		t.Run(tc.path, func(t *testing.T) {
			store := &contractStore{}
			handler, authCalls := contractTestHandler(t, store)
			w := httptest.NewRecorder()
			handler.ServeHTTP(w, contractRequest(tc.method, tc.path, tc.body))
			if w.Code != 400 || !strings.Contains(w.Body.String(), `"code":"`+tc.code+`"`) {
				t.Fatalf("unexpected error: %d %s", w.Code, w.Body.String())
			}
			if store.reads != 0 || store.writes != 0 || *authCalls != 1 {
				t.Fatalf("invalid request reached store or repeated authentication")
			}
			if strings.Contains(w.Body.String(), "sensitive") || strings.Contains(w.Body.String(), "secret-invalid") || strings.Contains(w.Body.String(), "private-token") {
				t.Fatal("validator exposed input")
			}
		})
	}
}

func TestContractPreservesAuthAndIdempotency(t *testing.T) {
	store := &contractStore{}
	handler, calls := contractTestHandler(t, store)
	req := contractRequest("POST", "/api/v1/attendance/corrections", `{"groupId":"g","recordId":"r1"}`)
	req.Header.Del("Authorization")
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)
	if w.Code != 401 || *calls != 0 || store.writes != 0 {
		t.Fatalf("unauthenticated request reached identity service or store")
	}
	key := "00000000-0000-4000-8000-000000000001"
	for _, conflict := range []bool{false, true} {
		store.conflict = conflict
		req = contractRequest("POST", "/api/v1/attendance/corrections", `{"groupId":"g","recordId":"r1","breaks":null}`)
		req.Header.Set("Idempotency-Key", key)
		w = httptest.NewRecorder()
		handler.ServeHTTP(w, req)
		want := 202
		if conflict {
			want = 409
		}
		if w.Code != want || store.key != key {
			t.Fatalf("idempotency changed: %d %s", w.Code, w.Body.String())
		}
	}
	req = contractRequest("POST", "/api/v1/attendance/corrections", `{"groupId":"g","recordId":"r1"}`)
	req.Header.Add("Idempotency-Key", key)
	req.Header.Add("Idempotency-Key", key)
	writes := store.writes
	w = httptest.NewRecorder()
	handler.ServeHTTP(w, req)
	if w.Code != 400 || store.writes != writes {
		t.Fatal("duplicate key was accepted")
	}
}

func TestAttendanceResponsesConformToContract(t *testing.T) {
	spec, err := httpapi.Spec()
	if err != nil {
		t.Fatal(err)
	}
	router, err := gorillamux.NewRouter(spec)
	if err != nil {
		t.Fatal(err)
	}
	store := &contractStore{}
	handler, _ := contractTestHandler(t, store)
	cases := []struct{ method, path, body string }{
		{"GET", "/api/v1/attendance/groups", ""},
		{"GET", "/api/v1/attendance/groups/page", ""},
		{"GET", "/api/v1/attendance/groups/page?pageSize=&cursor=", ""},
		{"GET", "/api/v1/attendance/records?group=g&month=2026-11", ""},
		{"GET", "/api/v1/attendance/records/r1?group=g", ""},
		{"GET", "/api/v1/attendance/monthly?group=g&month=2026-11", ""},
		{"GET", "/api/v1/attendance/monthly/summary?group=g&month=2026-11", ""},
		{"GET", "/api/v1/attendance/corrections/c1", ""},
		{"GET", "/api/v1/attendance/corrections/c1/review", ""},
		{"POST", "/api/v1/attendance/corrections", `{"groupId":"g","recordId":"r1"}`},
		{"POST", "/api/v1/attendance/corrections/c1/approve", `{}`},
		{"POST", "/api/v1/attendance/launch", `{"launch":"` + strings.Repeat("b", 64) + `"}`},
		{"POST", "/api/v1/attendance/punches", `{"requestId":"` + strings.Repeat("a", 64) + `","latitude":25,"longitude":121,"accuracyMeters":1,"capturedAt":"` + time.Now().UTC().Format(time.RFC3339) + `"}`},
	}
	for _, tc := range cases {
		t.Run(tc.path, func(t *testing.T) {
			req := contractRequest(tc.method, tc.path, tc.body)
			w := httptest.NewRecorder()
			handler.ServeHTTP(w, req)
			if w.Code >= 500 {
				t.Fatalf("handler failed: %d %s", w.Code, w.Body.String())
			}
			route, params, err := router.FindRoute(req)
			if err != nil {
				t.Fatal(err)
			}
			input := &openapi3filter.ResponseValidationInput{RequestValidationInput: &openapi3filter.RequestValidationInput{Request: req, PathParams: params, Route: route}, Status: w.Code, Header: w.Header(), Body: io.NopCloser(bytes.NewReader(w.Body.Bytes()))}
			if err := openapi3filter.ValidateResponse(context.Background(), input); err != nil {
				t.Fatalf("response violates contract: %v\n%s", err, w.Body.String())
			}
			if strings.Contains(tc.path, "records/r1") {
				var data map[string]any
				if err := json.Unmarshal(w.Body.Bytes(), &data); err != nil {
					t.Fatal(err)
				}
				if data["Breaks"] != nil || data["RebuildIssues"] != nil {
					t.Fatal("nullable arrays changed")
				}
				if _, ok := data["ClockInPosition"]; ok {
					t.Fatal("detail exposed location")
				}
			}
		})
	}
}

func TestContractDTOPreservesLegacyWireValues(t *testing.T) {
	now := time.Date(2026, 11, 1, 8, 30, 0, 123, time.UTC)
	for _, empty := range []bool{false, true} {
		record := attendance.Record{RecordID: "r1", GroupID: "g", ClockInAt: now, ClockOutAt: now.Add(time.Hour), ClockInPosition: &attendance.Position{Latitude: 25.1234567890123, Longitude: 121.9876543210123, AccuracyMeters: 1.123456789, CapturedAt: now}, RestConfirmed: true, Revision: 7, EventSequence: 9, SchemaVersion: 2}
		if empty {
			record.Breaks = []worktime.Interval{}
			record.RebuildIssues = []string{}
			record.SourceEventIDs = []string{}
		}
		report := attendance.CalculateMonthlyAttendance([]attendance.Record{record}, now)
		name := "nil"
		if empty {
			name = "empty"
		}
		original, err := os.ReadFile("testdata/legacy-monthly-" + name + ".json")
		if err != nil {
			t.Fatal(err)
		}
		converted, err := json.Marshal(monthlyView(report))
		if err != nil {
			t.Fatal(err)
		}
		var before, after any
		if err := json.Unmarshal(original, &before); err != nil {
			t.Fatal(err)
		}
		if err := json.Unmarshal(converted, &after); err != nil {
			t.Fatal(err)
		}
		if !reflect.DeepEqual(before, after) {
			t.Fatalf("DTO changed legacy JSON or location precision:\n%s\n%s", original, converted)
		}
	}
}
