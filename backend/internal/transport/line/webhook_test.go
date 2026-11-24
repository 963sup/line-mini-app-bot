package line

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	lineplatform "github.com/963sup/line-mini-app-bot/internal/platform/line"
	"github.com/line/line-bot-sdk-go/v8/linebot/messaging_api"
	"github.com/line/line-bot-sdk-go/v8/linebot/webhook"
)

type replyTransport func(*http.Request) (*http.Response, error)

func (f replyTransport) RoundTrip(r *http.Request) (*http.Response, error) { return f(r) }

func TestRedeliveredBatchContinuesAfterSpentReplyToken(t *testing.T) {
	sends := 0
	bot, err := lineplatform.NewMessaging("test", &http.Client{Transport: replyTransport(func(*http.Request) (*http.Response, error) {
		sends++
		status, body := 200, "{}"
		if sends == 2 {
			status, body = 400, `{"message":"Invalid reply token"}`
		}
		return &http.Response{StatusCode: status, Body: io.NopCloser(strings.NewReader(body))}, nil
	})})
	if err != nil {
		t.Fatal(err)
	}
	second := 0
	handler := WebhookHandler("secret", bot.Reply, Dispatch{Commands: []Handler{{RetryOnSendFailure: true, Handle: func(_ context.Context, event webhook.EventInterface) (Reply, bool, error) {
		e := event.(webhook.MessageEvent)
		if e.WebhookEventId == "second" {
			second++
			if second == 1 {
				return Reply{}, true, errors.New("temporary")
			}
		}
		return Reply{Token: e.ReplyToken, Message: &messaging_api.TextMessage{Text: "ok"}}, true, nil
	}}}})
	for i, want := range []int{503, 200} {
		event := func(id string) string {
			return fmt.Sprintf(`{"type":"message","message":{"type":"text","text":"hello"},"webhookEventId":%q,"replyToken":%q,"deliveryContext":{"isRedelivery":%t}}`, id, id, i == 1)
		}
		body := `{"events":[` + event("first") + "," + event("second") + "]}"
		mac := hmac.New(sha256.New, []byte("secret"))
		mac.Write([]byte(body))
		r := httptest.NewRequest("POST", "/", strings.NewReader(body))
		r.Header.Set("X-Line-Signature", base64.StdEncoding.EncodeToString(mac.Sum(nil)))
		w := httptest.NewRecorder()
		handler(w, r)
		if w.Code != want {
			t.Fatalf("delivery %d: got %d want %d", i, w.Code, want)
		}
	}
	if second != 2 || sends != 3 {
		t.Fatalf("second visits=%d, sends=%d", second, sends)
	}
}

func TestWebhookDispatchPrecedenceAndObservationFailure(t *testing.T) {
	bot, err := lineplatform.NewMessaging("test", http.DefaultClient)
	if err != nil {
		t.Fatal(err)
	}
	for _, tc := range []struct {
		name string
		want string
		code int
	}{
		{"early command", "early", http.StatusOK},
		{"observer failure", "early,observe", http.StatusServiceUnavailable},
		{"fallback command", "early,observe,command", http.StatusOK},
	} {
		t.Run(tc.name, func(t *testing.T) {
			var calls []string
			handler := WebhookHandler("secret", bot.Reply, Dispatch{
				EarlyCommands: []Handler{{Handle: func(context.Context, webhook.EventInterface) (Reply, bool, error) {
					calls = append(calls, "early")
					return Reply{}, tc.name == "early command", nil
				}}},
				Observers: []Observer{{Failure: "observation unavailable", Observe: func(_ context.Context, _ webhook.EventInterface, group *GroupEvent) error {
					if group == nil || group.EventID != "group-event" || group.GroupID != "C"+strings.Repeat("a", 32) {
						t.Fatalf("observer did not receive decoded group facts: %+v", group)
					}
					calls = append(calls, "observe")
					if tc.name == "observer failure" {
						return errors.New("temporary")
					}
					return nil
				}}},
				Commands: []Handler{{Handle: func(context.Context, webhook.EventInterface) (Reply, bool, error) {
					calls = append(calls, "command")
					return Reply{}, true, nil
				}}},
			})
			body := fmt.Sprintf(`{"events":[{"type":"message","mode":"active","timestamp":1000,"webhookEventId":"group-event","source":{"type":"group","groupId":%q},"message":{"type":"text","text":"hello"}}]}`, "C"+strings.Repeat("a", 32))
			mac := hmac.New(sha256.New, []byte("secret"))
			mac.Write([]byte(body))
			r := httptest.NewRequest("POST", "/", strings.NewReader(body))
			r.Header.Set("X-Line-Signature", base64.StdEncoding.EncodeToString(mac.Sum(nil)))
			w := httptest.NewRecorder()
			handler(w, r)
			if got := strings.Join(calls, ","); got != tc.want || w.Code != tc.code {
				t.Fatalf("got %q and status %d, want %q and status %d", got, w.Code, tc.want, tc.code)
			}
		})
	}
}
