package messaging

import (
	"context"
	"encoding/base64"
	"encoding/json"

	"github.com/963sup/line-mini-app-bot/internal/business/object/domain/event"
	"google.golang.org/api/pubsub/v1"
)

// Publisher delivers one immutable revision to the configured Pub/Sub topic.
type Publisher struct {
	Client *pubsub.Service
	Topic  string
}

func (p Publisher) Publish(ctx context.Context, id string, change event.Revision) error {
	data, err := json.Marshal(change)
	if err != nil {
		return err
	}
	_, err = p.Client.Projects.Topics.Publish(p.Topic, &pubsub.PublishRequest{Messages: []*pubsub.PubsubMessage{{Data: base64.StdEncoding.EncodeToString(data), Attributes: map[string]string{"schemaVersion": "1", "eventId": id, "type": "object.quantity_changed"}}}}).Context(ctx).Do()
	return err
}
