package line

import (
	"context"
	"strings"

	repositoryApplication "github.com/963sup/line-mini-app-bot/internal/business/repository/application"
	repositoryDomainEntity "github.com/963sup/line-mini-app-bot/internal/business/repository/domain/entity"
	transport "github.com/963sup/line-mini-app-bot/internal/transport/line"
	"github.com/963sup/line-mini-app-bot/internal/transport/line/navigation"
	"github.com/line/line-bot-sdk-go/v8/linebot/messaging_api"
	"github.com/line/line-bot-sdk-go/v8/linebot/webhook"
)

func EngineeringHandler(appURL string, catalog *repositoryApplication.Service) transport.Handler {
	return transport.Handler{Owner: "repository", Commands: []string{"我的工程"}, Failure: "repository temporarily unavailable", Handle: func(ctx context.Context, event webhook.EventInterface) (transport.Reply, bool, error) {
		return MyEngineering(ctx, event, appURL, catalog)
	}}
}

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
	page, err := catalog.List(ctx, user.UserId, "", "")
	if err != nil {
		return transport.Reply{}, true, err
	}
	columns := []messaging_api.CarouselColumn{}
	for _, repo := range page.Items[:min(len(page.Items), 5)] {
		name := repo.Name
		if name == "" {
			name = "工程儲存庫"
		}
		if link := navigation.RepositoryURL(appURL, true, repo.ID); link != "" {
			columns = append(columns, messaging_api.CarouselColumn{Title: "我的工程", Text: name, Actions: []messaging_api.ActionInterface{&messaging_api.UriAction{Label: "開啟工程", Uri: link}}})
		}
	}
	if len(page.Items) == 0 && page.NextCursor == "" {
		return transport.Reply{Token: msg.ReplyToken, Message: &messaging_api.TextMessage{Text: "目前沒有可開啟的工程。"}}, true, nil
	}
	if len(page.Items) > 5 || page.NextCursor != "" {
		if link := navigation.RepositoryURL(appURL, false); link != "" {
			columns = append(columns, messaging_api.CarouselColumn{Title: "我的工程", Text: "還有更多可開啟的工程", Actions: []messaging_api.ActionInterface{&messaging_api.UriAction{Label: "查看完整清單", Uri: link}}})
		}
	}
	if len(columns) == 0 {
		return transport.Reply{Token: msg.ReplyToken, Message: &messaging_api.TextMessage{Text: "工程入口目前無法使用。"}}, true, nil
	}
	return transport.Reply{Token: msg.ReplyToken, Message: &messaging_api.TemplateMessage{AltText: "選擇我的工程", Template: &messaging_api.CarouselTemplate{Columns: columns}}}, true, nil
}
