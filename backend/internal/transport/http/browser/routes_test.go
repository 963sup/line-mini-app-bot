package browser

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestRemovedWorkspacePreviewIsNotServed(t *testing.T) {
	mux := http.NewServeMux()
	if err := Register(mux); err != nil {
		t.Fatal(err)
	}
	for name, handler := range map[string]http.Handler{
		"registered routes": mux,
		"page dispatcher":   http.HandlerFunc(serveMiniApp),
	} {
		t.Run(name, func(t *testing.T) {
			recorder := httptest.NewRecorder()
			handler.ServeHTTP(recorder, httptest.NewRequest("GET", "/miniapp/workspace-preview", nil))
			if recorder.Code != http.StatusNotFound {
				t.Fatalf("removed preview returned HTTP %d, want 404", recorder.Code)
			}
		})
	}
}

func TestMiniAppAttendanceGetsDedicatedPage(t *testing.T) {
	for _, tc := range []struct {
		url        string
		attendance bool
		content    string
	}{
		{"/miniapp", false, `aria-label="首頁常用功能"`},
		{"/miniapp/feed", false, `"page":"/miniapp/feed"`},
		{"/miniapp?journal=1", false, `"page":"/miniapp/feed"`},
		{"/miniapp?liff.state=%3Fjournal%3D1", false, `"page":"/miniapp/feed"`},
		{"/miniapp?launch=token", true, "確認並完成打卡"},
		{"/miniapp?liff.state=%3Flaunch%3Dtoken", true, "確認並完成打卡"},
		{"/miniapp/attendance", false, "出勤紀錄"},
		{"/miniapp/leave", false, "請假登記"},
		{"/miniapp/correction", false, "申請補卡"},
		{"/miniapp/review", false, "補卡審核"},
	} {
		recorder := httptest.NewRecorder()
		serveMiniApp(recorder, httptest.NewRequest("GET", tc.url, nil))
		body := recorder.Body.String()
		if recorder.Code != 200 {
			t.Fatalf("%s status %d", tc.url, recorder.Code)
		}
		if !strings.Contains(recorder.Header().Get("Content-Security-Policy"), "frame-src https://access.line.me;") {
			t.Errorf("%s unexpected frame policy", tc.url)
		}
		if strings.Contains(body, `id="punch"`) != tc.attendance {
			t.Errorf("%s attendance page = %v, want %v", tc.url, strings.Contains(body, `id="punch"`), tc.attendance)
		}
		if !strings.Contains(body, tc.content) {
			t.Errorf("%s missing workflow content %s", tc.url, tc.content)
		}
		if tc.attendance && strings.Contains(body, `id="worktime"`) {
			t.Errorf("%s includes worktime UI", tc.url)
		}
	}
}

func TestObjectRecognitionToolRoute(t *testing.T) {
	mux := http.NewServeMux()
	if err := Register(mux); err != nil {
		t.Fatal(err)
	}
	response := httptest.NewRecorder()
	mux.ServeHTTP(response, httptest.NewRequest("GET", "/miniapp/object-recognition", nil))
	if response.Code != http.StatusOK || !strings.Contains(response.Body.String(), `"page":"/miniapp/object-recognition"`) {
		t.Fatalf("object recognition tool page not served (HTTP %d)", response.Code)
	}
}

func TestLegacyAttendanceLinksPreserveQuery(t *testing.T) {
	mux := http.NewServeMux()
	if err := Register(mux); err != nil {
		t.Fatal(err)
	}
	for source, target := range map[string]string{
		"/miniapp/history?month=2026-11":             "/miniapp/attendance?month=2026-11",
		"/miniapp/attendance?tab=correction&group=g": "/miniapp/correction?group=g",
		"/miniapp/attendance?tab=review&id=r":        "/miniapp/review?id=r",
		"/miniapp/attendance?tab=leave&leave=token":  "/miniapp/leave?leave=token",
		"/miniapp?liff.state=%3Fleave%3Dtoken":       "/miniapp/leave?liff.state=%3Fleave%3Dtoken",
	} {
		recorder := httptest.NewRecorder()
		mux.ServeHTTP(recorder, httptest.NewRequest("GET", source, nil))
		if recorder.Code != http.StatusTemporaryRedirect || recorder.Header().Get("Location") != target {
			t.Errorf("%s got %d %s, want 307 %s", source, recorder.Code, recorder.Header().Get("Location"), target)
		}
	}
}
