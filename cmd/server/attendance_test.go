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
}

func (s *attendanceStub) Punch(_ context.Context, group, user, event, action string, now time.Time, _ int64) (attendance.Result, error) {
	s.calls++
	s.group, s.user, s.event = group, user, event
	s.action = action
	return attendance.Result{Code: "clocked_in", Record: attendance.Record{ClockInAt: now}}, s.err
}
func (s *attendanceStub) Recent(_ context.Context, user string) ([]attendance.Record, error) {
	s.queriedUser = user
	return nil, s.err
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
			b := &attendanceBot{store: s, zone: zone}
			e := webhook.PostbackEvent{Mode: "active", ReplyToken: "reply", WebhookEventId: "event", Source: webhook.GroupSource{GroupId: "group", UserId: "actor"}, Postback: &webhook.PostbackContent{Data: tc.data}, DeliveryContext: &webhook.DeliveryContext{IsRedelivery: true}}
			_, handled, err := b.reply(context.Background(), e)
			if err != nil || handled != tc.handled || s.action != tc.action {
				t.Fatalf("handled=%v action=%s err=%v", handled, s.action, err)
			}
			if handled && (s.calls != 1 || s.user != "actor" || s.group != "group" || s.event != "event") {
				t.Fatal("must use signed event identity and scope")
			}
		})
	}
}

func TestAttendanceHistoryUsesDirectActor(t *testing.T) {
	s := &attendanceStub{}
	b := &attendanceBot{store: s}
	e := webhook.MessageEvent{Mode: "active", ReplyToken: "reply", Source: webhook.UserSource{UserId: "actor"}, Message: webhook.TextMessageContent{Text: "出勤紀錄"}}
	_, handled, err := b.reply(context.Background(), e)
	if err != nil || !handled || s.queriedUser != "actor" || s.calls != 0 {
		t.Fatal("history must read only the direct actor without approval")
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
			b := &attendanceBot{store: s, zone: zone}
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
	b := &attendanceBot{store: s}
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
}
