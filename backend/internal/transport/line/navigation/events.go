package navigation

import (
	"context"
	"net/url"
	"strings"

	transport "github.com/963sup/line-mini-app-bot/internal/transport/line"
	"github.com/line/line-bot-sdk-go/v8/linebot/messaging_api"
	"github.com/line/line-bot-sdk-go/v8/linebot/webhook"
)

type menuRequest struct {
	token   string
	command string
	groupID string
}

type Menu struct{ StickerPackageID, StickerID string }

func (trigger Menu) parseMenu(event webhook.EventInterface) menuRequest {
	var source webhook.SourceInterface
	var command, token string
	switch e := event.(type) {
	case webhook.JoinEvent:
		g, ok := e.Source.(webhook.GroupSource)
		if !ok || g.GroupId == "" || !transport.CanReply(e.Mode, e.ReplyToken, e.DeliveryContext) {
			return menuRequest{}
		}
		return menuRequest{token: e.ReplyToken, command: "/help", groupID: g.GroupId}
	case webhook.MessageEvent:
		if !transport.CanReply(e.Mode, e.ReplyToken, e.DeliveryContext) {
			return menuRequest{}
		}
		if sticker, ok := e.Message.(webhook.StickerMessageContent); ok {
			g, group := e.Source.(webhook.GroupSource)
			if !group || g.GroupId == "" || sticker.PackageId == "" || sticker.StickerId == "" ||
				sticker.PackageId != trigger.StickerPackageID || sticker.StickerId != trigger.StickerID {
				return menuRequest{}
			}
			return menuRequest{token: e.ReplyToken, command: "/menu", groupID: g.GroupId}
		}
		text, ok := e.Message.(webhook.TextMessageContent)
		if !ok {
			return menuRequest{}
		}
		source, token = e.Source, e.ReplyToken
		command = strings.ToLower(strings.TrimSpace(text.Text))
	case webhook.PostbackEvent:
		if !transport.CanReply(e.Mode, e.ReplyToken, e.DeliveryContext) || e.Postback == nil {
			return menuRequest{}
		}
		values, err := url.ParseQuery(e.Postback.Data)
		if err != nil || len(values) != 1 || len(values["action"]) != 1 {
			return menuRequest{}
		}
		switch values.Get("action") {
		case "help", "miniapp":
		case "clock-in", "transfer", "clock-out":
			if _, ok := e.Source.(webhook.UserSource); !ok {
				return menuRequest{}
			}
		default:
			return menuRequest{}
		}
		command = "/" + values.Get("action")
		source, token = e.Source, e.ReplyToken
	default:
		return menuRequest{}
	}

	var groupID string
	switch s := source.(type) {
	case webhook.UserSource:
		if s.UserId == "" {
			return menuRequest{}
		}
	case webhook.GroupSource:
		if s.GroupId == "" {
			return menuRequest{}
		}
		groupID = s.GroupId
	default:
		return menuRequest{}
	}

	switch command {
	case "選單", "功能":
		command = "/menu"
	case "/menu", "/miniapp", "/help", "/clock-in", "/transfer", "/clock-out":
	default:
		if groupID != "" {
			return menuRequest{}
		}
		command = "private-help"
	}
	return menuRequest{token: token, command: command, groupID: groupID}
}

func (m menuRequest) reply(appURL, endpointURL string, ids ...string) transport.Reply {
	reply := transport.Reply{Token: m.token}
	switch m.command {
	case "/menu":
		reply.Message = QuickMenu(m.groupID != "", appURL, ids...)
	case "/miniapp":
		reply.Message = MiniApp(appURL, endpointURL, ids...)
	case "/help":
		reply.Message = Help(m.groupID != "", appURL, endpointURL, ids...)
	case "/clock-in":
		reply.Message = &messaging_api.TextMessage{Text: "請在上班的群組輸入「打卡」，再選擇「上班」。"}
	case "/transfer":
		reply.Message = &messaging_api.TextMessage{Text: "請先確認支援作業是否仍計入原班次。需要結束原班次時，請回原群組操作。此入口不會結束或新增班次。"}
	case "/clock-out":
		reply.Message = &messaging_api.TextMessage{Text: "請在要結束班次的原群組輸入「打卡」，再選擇「下班」。可在這裡輸入「出勤紀錄」查詢本人紀錄。"}
	case "private-help":
		reply.Message = Help(false, appURL, endpointURL)
	}
	return reply
}

func (trigger Menu) Handler(appURL, endpointURL string, lookup func(context.Context, string) (string, error)) transport.Handler {
	return transport.Handler{Handle: func(ctx context.Context, event webhook.EventInterface) (transport.Reply, bool, error) {
		menu := trigger.parseMenu(event)
		if menu.groupID != "" && lookup != nil {
			id, err := lookup(ctx, menu.groupID)
			if err != nil {
				return transport.Reply{}, false, err
			}
			return menu.reply(appURL, endpointURL, id), true, nil
		}
		return menu.reply(appURL, endpointURL), true, nil
	}}
}
