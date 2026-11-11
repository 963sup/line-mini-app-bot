package line

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/963sup/line-mini-app-bot/internal/business/weather/forecast"
	"github.com/line/line-bot-sdk-go/v8/linebot/webhook"
)

type weatherStub struct {
	calls    int
	lat, lon float64
	err      error
}

func (s *weatherStub) Lookup(_ context.Context, lat, lon float64, _ string) (forecast.Forecast, error) {
	s.calls++
	s.lat, s.lon = lat, lon
	return forecast.Forecast{Region: "臺北市", Description: "多雲", Start: time.Now(), End: time.Now().Add(time.Hour)}, s.err
}

func TestWeatherOnlyGroupLocation(t *testing.T) {
	for _, tc := range []struct {
		name       string
		source     webhook.SourceInterface
		message    webhook.MessageContentInterface
		redelivery bool
		want       bool
	}{
		{"LINE map", webhook.GroupSource{GroupId: "group"}, webhook.LocationMessageContent{Latitude: 25, Longitude: 121}, false, true},
		{"map URL", webhook.GroupSource{GroupId: "group"}, webhook.TextMessageContent{Text: "https://maps.google.com"}, false, false},
		{"direct location", webhook.UserSource{UserId: "user"}, webhook.LocationMessageContent{}, false, false},
		{"redelivery", webhook.GroupSource{GroupId: "group"}, webhook.LocationMessageContent{}, true, false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			s := &weatherStub{}
			e := webhook.MessageEvent{Mode: "active", ReplyToken: "reply", Source: tc.source, Message: tc.message, DeliveryContext: &webhook.DeliveryContext{IsRedelivery: tc.redelivery}}
			r, handled := ReplyForEvent(context.Background(), e, s)
			if handled != tc.want || (s.calls == 1) != tc.want {
				t.Fatalf("handled=%v calls=%d", handled, s.calls)
			}
			if handled && (r.Message == nil || s.lat != 25 || s.lon != 121) {
				t.Fatal("must query shared coordinates and reply")
			}
		})
	}
}

func TestWeatherFailureReplies(t *testing.T) {
	e := webhook.MessageEvent{Mode: "active", ReplyToken: "reply", Source: webhook.GroupSource{GroupId: "group"}, Message: webhook.LocationMessageContent{Latitude: 25, Longitude: 121}}
	for _, client := range []ForecastLookup{nil, &weatherStub{err: errors.New("upstream failure")}} {
		r, handled := ReplyForEvent(context.Background(), e, client)
		if !handled || r.Message == nil {
			t.Fatal("configuration and upstream failure must get a short reply")
		}
	}
}
