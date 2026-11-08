package main

import (
	"context"
	"errors"
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

func (s *attendanceStub) CreateRequest(_ context.Context, request attendance.Request) (string, error) {
	s.calls++
	s.group, s.user, s.event, s.action = request.GroupID, request.UserID, request.EventID, request.Action
	return strings.Repeat("a", 64), s.err
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
			b := &attendanceBot{store: struct{ attendance.Store }{s}, zone: zone}
			e := webhook.MessageEvent{Mode: "active", ReplyToken: "reply", Source: webhook.UserSource{UserId: "actor"}, Message: webhook.TextMessageContent{Text: "出勤紀錄"}}
			reply, handled, err := b.reply(context.Background(), e)
			text, ok := reply.message.(*messaging_api.TextMessage)
			if err != nil || !handled || !ok || text.Text != tc.want || s.queriedUser != "actor" || s.calls != 0 {
				t.Fatalf("unexpected history response: message=%+v handled=%v err=%v", reply.message, handled, err)
			}
		})
	}
}

func TestAttendancePostback(t *testing.T) {
	zone, _ := time.LoadLocation("Asia/Taipei")
	for _, tc := range []struct {
		data, action string
		handled      bool
	}{
		{"action=clock-in", "clock-in", true},
		{"action=clock-out", "clock-out", true},
		{"action=clock-in&userId=other", "", false},
		{"action=clock-in&action=clock-out", "", false},
		{"action=help", "", false},
		{"action=%zz", "", false},
	} {
		t.Run(tc.data, func(t *testing.T) {
			s := &attendanceStub{}
			b := &attendanceBot{store: s, zone: zone, appURL: "https://liff.line.me/test"}
			e := webhook.PostbackEvent{Mode: "active", ReplyToken: "reply", WebhookEventId: "event", Source: webhook.GroupSource{GroupId: "group", UserId: "actor"}, Postback: &webhook.PostbackContent{Data: tc.data}, DeliveryContext: &webhook.DeliveryContext{IsRedelivery: true}}
			reply, handled, err := b.reply(context.Background(), e)
			if err != nil || handled != tc.handled || s.action != tc.action {
				t.Fatalf("handled=%v action=%s err=%v", handled, s.action, err)
			}
			if handled && (s.calls != 1 || s.user != "actor" || s.group != "group" || s.event != "event") {
				t.Fatal("must use signed event identity and scope")
			}
			if s.punchCalls != 0 {
				t.Fatal("group click must not punch without position")
			}
			if handled {
				if _, ok := reply.message.(*messaging_api.TemplateMessage); !ok {
					t.Fatal("must provide location link")
				}
			}
		})
	}
}

func TestAttendanceHistoryUsesDirectActor(t *testing.T) {
	s := &attendanceStub{groups: []string{"group-a", "group-b"}}
	b := &attendanceBot{store: s}
	e := webhook.MessageEvent{Mode: "active", ReplyToken: "reply", Source: webhook.UserSource{UserId: "actor"}, Message: webhook.TextMessageContent{Text: "出勤紀錄"}}
	_, handled, err := b.reply(context.Background(), e)
	if err != nil || !handled || s.queriedUser != "actor" || s.calls != 0 {
		t.Fatal("history must read only the direct actor without approval")
	}
	e.Message = webhook.TextMessageContent{Text: "出勤紀錄 2"}
	_, handled, err = b.reply(context.Background(), e)
	if err != nil || !handled || s.queriedGroup != "group-b" || s.queriedUser != "actor" || s.calls != 0 {
		t.Fatal("history must stay in selected group and direct actor")
	}
}

func TestAttendanceMenuIsFlexWithoutWrite(t *testing.T) {
	s := &attendanceStub{}
	b := &attendanceBot{store: s}
	e := webhook.MessageEvent{Mode: "active", ReplyToken: "reply", Source: webhook.GroupSource{GroupId: "group", UserId: "actor"}, Message: webhook.TextMessageContent{Text: "打卡"}}
	reply, handled, err := b.reply(context.Background(), e)
	if _, ok := reply.message.(*messaging_api.FlexMessage); !ok || !handled || err != nil || s.calls != 0 {
		t.Fatal("menu must return a Flex card without punching")
	}
}

func TestAttendanceIdentityAndRedelivery(t *testing.T) {
	zone, _ := time.LoadLocation("Asia/Taipei")
	for _, tc := range []struct {
		name, group, user, event, command string
		redelivery                        bool
		calls                             int
	}{
		{"group member", "g", "u", "event", "上班", false, 1},
		{"redelivery reaches deduplicating store", "g", "u", "event", "上班", true, 1},
		{"another group needs no approval", "other", "u", "event", "上班", false, 1},
		{"another member needs no approval", "g", "other", "event", "上班", false, 1},
		{"missing group", "", "u", "event", "上班", false, 0},
		{"menu does not punch", "g", "u", "event", "打卡", false, 0},
		{"missing actor", "g", "", "event", "上班", false, 0},
		{"missing event", "g", "u", "", "上班", false, 0},
		{"history stays private", "g", "u", "event", "出勤紀錄", false, 0},
		{"ordinary conversation", "g", "u", "event", "你好", false, 0},
	} {
		t.Run(tc.name, func(t *testing.T) {
			s := &attendanceStub{}
			b := &attendanceBot{store: s, zone: zone, appURL: "https://liff.line.me/test"}
			e := webhook.MessageEvent{Mode: "active", ReplyToken: "token", WebhookEventId: tc.event, Source: webhook.GroupSource{GroupId: tc.group, UserId: tc.user}, Message: webhook.TextMessageContent{Text: tc.command}, DeliveryContext: &webhook.DeliveryContext{IsRedelivery: tc.redelivery}}
			_, _, err := b.reply(context.Background(), e)
			if err != nil || s.calls != tc.calls {
				t.Fatalf("calls=%d error=%v", s.calls, err)
			}
			if s.calls > 0 && (s.group != tc.group || s.user != tc.user || s.event != tc.event) {
				t.Fatal("actor or scope changed")
			}
		})
	}
}

func TestAttendanceFailureDoesNotReplySuccess(t *testing.T) {
	s := &attendanceStub{err: errors.New("unavailable")}
	b := &attendanceBot{store: s, appURL: "https://liff.line.me/test"}
	e := webhook.MessageEvent{Mode: "active", ReplyToken: "token", WebhookEventId: "event", Source: webhook.GroupSource{GroupId: "g", UserId: "u"}, Message: webhook.TextMessageContent{Text: "上班"}}
	reply, handled, err := b.reply(context.Background(), e)
	if !handled || err == nil || reply.message != nil {
		t.Fatal("must not acknowledge an unsuccessful write")
	}
}

func TestAttendanceWebhookRejectsUnsignedRequest(t *testing.T) {
	s := &attendanceStub{}
	b := &attendanceBot{store: s}
	r := httptest.NewRequest(http.MethodPost, "/webhooks/line/123", strings.NewReader(`{"events":[]}`))
	w := httptest.NewRecorder()
	lineWebhookHandler("secret", nil, "", b, nil)(w, r)
	if w.Code != http.StatusUnauthorized || s.calls != 0 {
		t.Fatalf("status=%d calls=%d", w.Code, s.calls)
	}
	t.Setenv("LINE_MINI_APP_CHANNEL_ID", "123")
	mux := http.NewServeMux()
	registerAttendanceAPI(mux, b)
	r = httptest.NewRequest(http.MethodPost, "/api/v1/attendance/punches", strings.NewReader(`{}`))
	w = httptest.NewRecorder()
	mux.ServeHTTP(w, r)
	if w.Code != http.StatusUnauthorized || s.calls != 0 {
		t.Fatal("browser punch must require LINE authentication")
	}
}
