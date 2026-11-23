package messaging

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/963sup/line-mini-app-bot/internal/business/object/domain/event"
	"google.golang.org/api/option"
	"google.golang.org/api/pubsub/v1"
)

func TestPublishQuantityRevision(t *testing.T) {
	revision := event.Revision{ChannelID: "channel", ObjectID: "object", Version: 2}
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" || r.URL.Path != "/v1/projects/example/topics/events:publish" {
			t.Errorf("unexpected request: %s %s", r.Method, r.URL.Path)
		}
		var request pubsub.PublishRequest
		if err := json.NewDecoder(r.Body).Decode(&request); err != nil || len(request.Messages) != 1 {
			t.Error("invalid message", err)
			w.WriteHeader(400)
			return
		}
		msg := request.Messages[0]
		payload, err := base64.StdEncoding.DecodeString(msg.Data)
		var got event.Revision
		if err != nil || json.Unmarshal(payload, &got) != nil || got != revision || msg.Attributes["eventId"] != "revision-id" {
			t.Error("incorrect revision", got)
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"messageIds":["1"]}`))
	}))
	defer server.Close()
	ctx := context.Background()
	client, err := pubsub.NewService(ctx, option.WithEndpoint(server.URL+"/"), option.WithoutAuthentication())
	if err != nil {
		t.Fatal(err)
	}
	if err := (Publisher{Client: client, Topic: "projects/example/topics/events"}).Publish(ctx, "revision-id", revision); err != nil {
		t.Fatal(err)
	}
}
