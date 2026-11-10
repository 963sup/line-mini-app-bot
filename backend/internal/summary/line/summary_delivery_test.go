package line

import (
	"context"
	"errors"
	"io"
	"net/http"
	"strings"
	"testing"

	"github.com/963sup/line-mini-app-bot/internal/summary"
	"github.com/line/line-bot-sdk-go/v8/linebot/messaging_api"
)

type deliveryTransport func(*http.Request) (*http.Response, error)

func (f deliveryTransport) RoundTrip(r *http.Request) (*http.Response, error) { return f(r) }
func TestDeliveryRetainsRetryAndHidesSDKFailures(t *testing.T) {
	for _, code := range []int{200, 409, 500} {
		bot, err := messaging_api.NewMessagingApiAPI("test-token", messaging_api.WithHTTPClient(&http.Client{Transport: deliveryTransport(func(r *http.Request) (*http.Response, error) {
			if r.Header.Get("X-Line-Retry-Key") != "stable-key" {
				t.Error("retry key changed")
			}
			return &http.Response{StatusCode: code, Header: make(http.Header), Body: io.NopCloser(strings.NewReader(`{"message":"private provider details"}`))}, nil
		})}))
		if err != nil {
			t.Fatal(err)
		}
		err = (SummaryDelivery{Bot: bot}).Send(context.Background(), "group", "text", "stable-key")
		if code == 500 {
			if !errors.Is(err, summary.ErrDeliveryUnavailable) || strings.Contains(err.Error(), "private") {
				t.Fatal("SDK failure escaped output port", err)
			}
		} else if err != nil {
			t.Fatal("accepted delivery became a failure", err)
		}
	}
}
