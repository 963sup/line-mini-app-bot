package browser

import (
	"net/http/httptest"
	"strings"
	"testing"
)

func TestMiniAppAttendanceGetsDedicatedPage(t *testing.T) {
	for _, tc := range []struct {
		url        string
		attendance bool
		content    string
	}{
		{"/miniapp", false, "選擇要完成的事"},
		{"/miniapp/journal", false, "日誌 Google 表單"},
		{"/miniapp?journal=1", false, "日誌 Google 表單"},
		{"/miniapp?liff.state=%3Fjournal%3D1", false, "日誌 Google 表單"},
		{"/miniapp?launch=token", true, "確認並完成打卡"},
		{"/miniapp?liff.state=%3Flaunch%3Dtoken", true, "確認並完成打卡"},
		{"/miniapp/history?group=g&month=2026-11", false, `id="record-list"`},
		{"/miniapp/correction?group=g&record=r", false, `id="correction-form"`},
		{"/miniapp/review", false, `id="review-details"`},
	} {
		recorder := httptest.NewRecorder()
		serveMiniApp(recorder, httptest.NewRequest("GET", tc.url, nil))
		body := recorder.Body.String()
		journal := tc.content == "日誌 Google 表單"
		if recorder.Code != 200 {
			t.Fatalf("%s status %d", tc.url, recorder.Code)
		}
		if strings.Contains(recorder.Header().Get("Content-Security-Policy"), "https://docs.google.com") != journal {
			t.Errorf("%s unexpected Google frame policy", tc.url)
		}
		if journal && !strings.Contains(body, "viewform?embedded=true") {
			t.Errorf("%s missing embedded form", tc.url)
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
