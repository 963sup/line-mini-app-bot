package line

import (
	"context"
	"strings"

	repositoryApplication "github.com/963sup/line-mini-app-bot/internal/business/repository/application"
	repositoryDomainEntity "github.com/963sup/line-mini-app-bot/internal/business/repository/domain/entity"
	"github.com/963sup/line-mini-app-bot/internal/transport/line/navigation"
	transport "github.com/963sup/line-mini-app-bot/internal/transport/line"
	"github.com/line/line-bot-sdk-go/v8/linebot/messaging_api"
	"github.com/line/line-bot-sdk-go/v8/linebot/webhook"
)

func MyEngineering(ctx context.Context, event webhook.EventInterface, appURL string, catalog *repositoryApplication.Service) (transport.Reply, bool, error) {
	msg, ok := event.(webhook.MessageEvent)
	if !ok || !transport.CanReply(msg.Mode, msg.ReplyToken, msg.DeliveryContext) {
		return transport.Reply{}, false, nil
	}
	user, private := msg.Source.(webhook.UserSource)
	text, isText := msg.Message.(webhook.TextMessageContent)
	if !private || user.UserId == "" || !isText || strings.TrimSpace(text.Text) != "我的工程" {
		return transport.Reply{}, false, nil
	}
	if catalog == nil {
		return transport.Reply{}, true, repositoryDomainEntity.ErrUnavailable
	}
	page, err := catalog.List(ctx, user.UserId, "")
	if err != nil {
		return transport.Reply{}, true, err
	}
	lines := []string{"我的工程"}
	for _, repo := range page.Items[:min(len(page.Items), 5)] {
		name := repo.Name
		if name == "" {
			name = "工程儲存庫"
		}
		if link := navigation.RepositoryURL(appURL, true, repo.ID); link != "" {
			name += "\n" + link
		}
		lines = append(lines, name)
	}
	if len(page.Items) == 0 && page.NextCursor == "" {
		lines = append(lines, "目前沒有可開啟的工程。")
	}
	if len(page.Items) > 5 || page.NextCursor != "" {
		if link := navigation.RepositoryURL(appURL, false); link != "" {
			lines = append(lines, "查看完整清單："+link)
		}
	}
	return transport.Reply{Token: msg.ReplyToken, Message: &messaging_api.TextMessage{Text: strings.Join(lines, "\n\n")}}, true, nil
}
