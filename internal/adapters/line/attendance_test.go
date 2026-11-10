package line

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/963sup/line-mini-app-bot/internal/attendance"
	"github.com/line/line-bot-sdk-go/v8/linebot/messaging_api"
	"github.com/line/line-bot-sdk-go/v8/linebot/webhook"
)

type attendanceStub struct {
	calls              int
	err                error
	group, user, event string
	action             string
	queriedUser        string
	queriedGroup       string
	groups             []string
	records            []attendance.Record
	punchCalls         int
}

func (s *attendanceStub) Submit(context.Context, string, string, attendance.Position, time.Time) (attendance.Result, error) {
	return attendance.Result{}, s.err
}

func (s *attendanceStub) Groups(_ context.Context, user string) ([]string, error) {
	s.queriedUser = user
	return s.groups, s.err
}

func (s *attendanceStub) RecentGroup(_ context.Context, user, group string) ([]attendance.Record, error) {
	s.queriedUser = user
	s.queriedGroup = group
	return nil, s.err
}

func (s *attendanceStub) Punch(_ context.Context, group, user, event, action string, now time.Time, _ int64) (attendance.Result, error) {
	s.punchCalls++
	s.calls++
	s.group, s.user, s.event = group, user, event
	s.action = action
	return attendance.Result{Code: "clocked_in", Record: attendance.Record{ClockInAt: now}}, s.err
}

func (s *attendanceStub) Recent(_ context.Context, user string) ([]attendance.Record, error) {
	s.queriedUser = user
	return s.records, s.err
}

func TestAttendanceLegacyHistory(t *testing.T) {
	zone, err := time.LoadLocation("Asia/Taipei")
	if err != nil {
		t.Fatal(err)
	}
	for _, tc := range []struct {
		name    string
		records []attendance.Record
		want    string
	}{
		{"empty", nil, "目前沒有出勤紀錄。"},
		{"active shift", []attendance.Record{{GroupID: "g", ClockInAt: time.Date(2026, 1, 2, 0, 0, 0, 0, time.UTC)}}, "本人最近出勤紀錄（最多五筆）：\n群組 g\n上班 2026-01-02 08:00:00\n下班 尚未下班"},
		{"completed shift", []attendance.Record{{GroupID: "g", ClockInAt: time.Date(2026, 1, 2, 0, 0, 0, 0, time.UTC), ClockOutAt: time.Date(2026, 1, 2, 9, 0, 0, 0, time.UTC)}}, "本人最近出勤紀錄（最多五筆）：\n群組 g\n上班 2026-01-02 08:00:00\n下班 2026-01-02 17:00:00"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			s := &attendanceStub{records: tc.records}
			// Expose the legacy Store contract without the optional LocatedStore methods.
			b := &Attendance{Store: struct{ attendance.Store }{s}, Zone: zone}
			e := webhook.MessageEvent{Mode: "active", ReplyToken: "reply", Source: webhook.UserSource{UserId: "actor"}, Message: webhook.TextMessageContent{Text: "出勤紀錄"}}
			reply, handled, err := b.reply(context.Background(), e)
			text, ok := reply.message.(*messaging_api.TextMessage)
			if err != nil || !handled || !ok || text.Text != tc.want || s.queriedUser != "actor" || s.calls != 0 {
				t.Fatalf("unexpected history response: message=%+v handled=%v err=%v", reply.message, handled, err)
			}
		})
	}
}

func TestOldActionPostbacksAreIgnored(t *testing.T) {
	b := &Attendance{Store: &attendanceStub{}, AppURL: "https://liff.line.me/test"}
	for _, action := range []string{"clock-in", "clock-out", "break-start", "break-end", "confirm-rest"} {
		e := webhook.PostbackEvent{Mode: "active", ReplyToken: "reply", Source: webhook.GroupSource{GroupId: "group", UserId: "actor"}, Postback: &webhook.PostbackContent{Data: "action=" + action}}
		reply, handled, err := b.reply(context.Background(), e)
		if err != nil || handled || reply.message != nil {
			t.Fatalf("old postback %s must be ignored", action)
		}
	}
}

func TestAttendanceHistoryUsesDirectActor(t *testing.T) {
	s := &attendanceStub{groups: []string{"group-a", "group-b"}}
	b := &Attendance{Store: s}
	e := webhook.MessageEvent{Mode: "active", ReplyToken: "reply", Source: webhook.UserSource{UserId: "actor"}, Message: webhook.TextMessageContent{Text: "出勤紀錄"}}
	_, handled, err := b.reply(context.Background(), e)
	if err != nil || !handled || s.queriedUser != "actor" || s.calls != 0 {
		t.Fatal("history must read only the direct actor without approval")
	}
	e.Message = webhook.TextMessageContent{Text: "出勤紀錄 2"}
	_, handled, err = b.reply(context.Background(), e)
	if err != nil || !handled || s.queriedGroup != "" || s.queriedUser != "actor" || s.calls != 0 {
		t.Fatal("history must stay in selected group and direct actor")
	}
}

func TestAttendanceWebhookRejectsUnsignedRequest(t *testing.T) {
	s := &attendanceStub{}
	b := &Attendance{Store: s}
	r := httptest.NewRequest(http.MethodPost, "/webhooks/line/123", strings.NewReader(`{"events":[]}`))
	w := httptest.NewRecorder()
	lineWebhookHandler("secret", nil, "", b, nil)(w, r)
	if w.Code != http.StatusUnauthorized || s.calls != 0 {
		t.Fatalf("status=%d calls=%d", w.Code, s.calls)
	}
}

func TestHistorySingleGroupAndPostbackAuthorization(t *testing.T) {
	s := &attendanceStub{groups: []string{"group-a"}}
	b := &Attendance{Store: s}
	e := webhook.MessageEvent{Mode: "active", ReplyToken: "reply", Source: webhook.UserSource{UserId: "actor"}, Message: webhook.TextMessageContent{Text: "出勤紀錄"}}
	reply, handled, err := b.reply(context.Background(), e)
	if err != nil || !handled || s.queriedGroup != "" {
		t.Fatalf("single group: %v %v %s", err, handled, s.queriedGroup)
	}
	if _, ok := reply.message.(*messaging_api.FlexMessage); !ok {
		t.Fatal("expected result card")
	}
	for _, group := range []string{"group-a", "forbidden"} {
		s.queriedGroup = ""
		p := webhook.PostbackEvent{Mode: "active", ReplyToken: "reply", Source: webhook.UserSource{UserId: "actor"}, Postback: &webhook.PostbackContent{Data: "action=attendance-history&month=2026-11&group=" + group}}
		_, handled, err = b.reply(context.Background(), p)
		if err != nil || !handled {
			t.Fatalf("postback %s: %v", group, err)
		}
		if group == "forbidden" && s.queriedGroup != "" {
			t.Fatal("queried unauthorized group")
		}
		if group == "group-a" && s.queriedGroup != group {
			t.Fatal("did not query selected group")
		}
	}
}

func (s *attendanceStub) MonthGroup(_ context.Context, user, group string, _ time.Time) ([]attendance.Record, error) {
	s.queriedUser = user
	s.queriedGroup = group
	return s.records, s.err
}
