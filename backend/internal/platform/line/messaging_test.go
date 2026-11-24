package line

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"strings"
	"testing"

	"github.com/line/line-bot-sdk-go/v8/linebot/messaging_api"
)

type messagingRoundTrip func(*http.Request) (*http.Response, error)

func (f messagingRoundTrip) RoundTrip(r *http.Request) (*http.Response, error) { return f(r) }

func TestSharedMessagingUsesEachRequestContext(t *testing.T) {
	type key struct{}
	client, err := NewMessaging("test-token", &http.Client{Transport: messagingRoundTrip(func(r *http.Request) (*http.Response, error) {
		if want := r.Context().Value(key{}); want != r.URL.Path {
			return nil, fmt.Errorf("unexpected context %v for %s", want, r.URL.Path)
		}
		return &http.Response{StatusCode: 200, Header: http.Header{}, Body: io.NopCloser(strings.NewReader("{}"))}, nil
	})})
	if err != nil {
		t.Fatal(err)
	}
	start := make(chan struct{})
	result := make(chan error, 2)
	go func() {
		<-start
		result <- client.PushText(context.WithValue(t.Context(), key{}, "/v2/bot/message/push"), "group", "hello", "retry")
	}()
	go func() {
		<-start
		result <- client.Reply(context.WithValue(t.Context(), key{}, "/v2/bot/message/reply"), "token", &messaging_api.TextMessage{Text: "hello"})
	}()
	close(start)
	for range 2 {
		if err := <-result; err != nil {
			t.Error(err)
		}
	}
}
