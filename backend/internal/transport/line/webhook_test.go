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

	"github.com/line/line-bot-sdk-go/v8/linebot/messaging_api"
	"github.com/line/line-bot-sdk-go/v8/linebot/webhook"
)

type replyTransport func(*http.Request) (*http.Response, error)

func (f replyTransport) RoundTrip(r *http.Request) (*http.Response, error) { return f(r) }

func TestRedeliveredBatchContinuesAfterSpentReplyToken(t *testing.T) {
	sends := 0
	bot, err := messaging_api.NewMessagingApiAPI("test", messaging_api.WithHTTPClient(&http.Client{Transport: replyTransport(func(*http.Request) (*http.Response, error) {
		sends++
		status, body := 200, "{}"
		if sends == 2 {
			status, body = 400, `{"message":"Invalid reply token"}`
		}
		return &http.Response{StatusCode: status, Body: io.NopCloser(strings.NewReader(body))}, nil
	})}))
	if err != nil {
		t.Fatal(err)
	}
	second := 0
	handler := WebhookHandler("secret", bot, []Handler{{RetryOnSendFailure: true, Handle: func(_ context.Context, event webhook.EventInterface) (Reply, bool, error) {
		e := event.(webhook.MessageEvent)
		if e.WebhookEventId == "second" {
			second++
			if second == 1 {
				return Reply{}, true, errors.New("temporary")
			}
		}
		return Reply{Token: e.ReplyToken, Message: &messaging_api.TextMessage{Text: "ok"}}, true, nil
	}}})
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
