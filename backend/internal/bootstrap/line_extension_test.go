package bootstrap

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"net/http"
	"net/http/httptest"
	"slices"
	"strings"
	"testing"

	channel "github.com/963sup/line-mini-app-bot/internal/transport/line"
	"github.com/line/line-bot-sdk-go/v8/linebot/messaging_api"
	"github.com/line/line-bot-sdk-go/v8/linebot/webhook"
)

func TestAdditionalCapabilityAtSignedWebhook(t *testing.T) {
	h := channel.Handler{Owner: "example", Commands: []string{"/example", "/ex"}, Postbacks: []string{"example-open"}}
	h.Handle = func(_ context.Context, event webhook.EventInterface) (channel.Reply, bool, error) {
		var token string
		switch e := event.(type) {
		case webhook.MessageEvent:
			text, ok := e.Message.(webhook.TextMessageContent)
			if !ok || !slices.Contains(h.Commands, text.Text) {
				return channel.Reply{}, false, nil
			}
			token = e.ReplyToken
		case webhook.PostbackEvent:
			if e.Postback == nil {
				return channel.Reply{}, false, nil
			}
			query, ok := channel.ParsePostback(e.Postback.Data)
			if !ok || !slices.Contains(h.Postbacks, query.Get("action")) {
				return channel.Reply{}, false, nil
			}
			token = e.ReplyToken
		default:
			return channel.Reply{}, false, nil
		}
		return channel.Reply{Token: token, Message: &messaging_api.TextMessage{Text: "example"}}, true, nil
	}
	mux := http.NewServeMux()
	var tokens []string
	channel.RegisterWebhook(mux, "example", "test-secret", func(_ context.Context, token string, message messaging_api.MessageInterface) error {
		if message.(*messaging_api.TextMessage).Text != "example" {
			t.Fatal(message)
		}
		tokens = append(tokens, token)
		return nil
	}, channel.Dispatch{Commands: []channel.Handler{h}})
	body := `{"events":[
		{"type":"message","message":{"type":"text","text":"/example"},"replyToken":"command"},
		{"type":"message","message":{"type":"text","text":"/ex"},"replyToken":"alias"},
		{"type":"postback","postback":{"data":"action=example-open"},"replyToken":"postback"},
		{"type":"message","message":{"type":"text","text":"other"},"replyToken":"unknown"}]}`
	mac := hmac.New(sha256.New, []byte("test-secret"))
	mac.Write([]byte(body))
	request := httptest.NewRequest(http.MethodPost, "/webhooks/line/example", strings.NewReader(body))
	request.Header.Set("X-Line-Signature", base64.StdEncoding.EncodeToString(mac.Sum(nil)))
	response := httptest.NewRecorder()
	mux.ServeHTTP(response, request)
	if response.Code != http.StatusOK || !slices.Equal(tokens, []string{"command", "alias", "postback"}) {
		t.Fatalf("status=%d reply tokens=%v", response.Code, tokens)
	}
}
