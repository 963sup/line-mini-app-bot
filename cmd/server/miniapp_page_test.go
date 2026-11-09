package main

import (
	"net/http/httptest"
	"strings"
	"testing"
)

func TestMiniAppAttendanceGetsDedicatedPage(t *testing.T) {
	for _, tc := range []struct {
		url        string
		attendance bool
	}{
		{"/miniapp", false},
		{"/miniapp?launch=token", true},
		{"/miniapp?liff.state=%3Flaunch%3Dtoken", true},
	} {
		recorder := httptest.NewRecorder()
		serveMiniApp(recorder, httptest.NewRequest("GET", tc.url, nil))
		body := recorder.Body.String()
		if strings.Contains(body, `id="punch"`) != tc.attendance {
			t.Errorf("%s attendance page = %v, want %v", tc.url, strings.Contains(body, `id="punch"`), tc.attendance)
		}
		if tc.attendance && strings.Contains(body, `id="worktime"`) {
			t.Errorf("%s includes worktime UI", tc.url)
		}
	}
}
