package main

import (
	"net/url"
	"strings"

	"github.com/963sup/line-mini-app-bot/internal/lineui"
	"github.com/line/line-bot-sdk-go/v8/linebot/messaging_api"
	"github.com/line/line-bot-sdk-go/v8/linebot/webhook"
)

type lineReply struct {
	token   string
	message messaging_api.MessageInterface
}

func replyForEvent(event webhook.EventInterface, appURL string) lineReply {
	var source webhook.SourceInterface
	var command, token string
	switch e := event.(type) {
	case webhook.MessageEvent:
		if !canReply(e.Mode, e.ReplyToken, e.DeliveryContext) {
			return lineReply{}
		}
		text, ok := e.Message.(webhook.TextMessageContent)
		if !ok {
			return lineReply{}
		}
		source, token = e.Source, e.ReplyToken
		command = strings.ToLower(strings.TrimSpace(text.Text))
	case webhook.PostbackEvent:
		if !canReply(e.Mode, e.ReplyToken, e.DeliveryContext) || e.Postback == nil {
			return lineReply{}
		}
		values, err := url.ParseQuery(e.Postback.Data)
		if err != nil || len(values) != 1 || len(values["action"]) != 1 {
			return lineReply{}
		}
		switch values.Get("action") {
		case "help":
			command = "/help"
		case "status":
			command = "/status"
		default:
			return lineReply{}
		}
		source, token = e.Source, e.ReplyToken
	default:
		return lineReply{}
	}

	group := false
	switch s := source.(type) {
	case webhook.UserSource:
		if s.UserId == "" {
			return lineReply{}
		}
	case webhook.GroupSource:
		if s.GroupId == "" {
			return lineReply{}
		}
		group = true
	default:
		return lineReply{}
	}

	switch command {
	case "/help", "/menu", "選單", "功能":
		return lineReply{token: token, message: lineui.Help(group, appURL)}
	case "/status":
		text := "一對一連線已接通；出勤尚未啟用。"
		if group {
			text = "群組連線已接通；此處只提供公開指令，不讀取個人資料。"
		}
		return lineReply{token: token, message: &messaging_api.TextMessage{Text: text}}
	default:
		// Ordinary group conversation must not trigger the bot or expose personal data.
		if group {
			return lineReply{}
		}
		return lineReply{token: token, message: lineui.Help(false, appURL)}
	}
}

func canReply(mode webhook.EventMode, token string, delivery *webhook.DeliveryContext) bool {
	if mode != "active" || token == "" {
		return false
	}
	return delivery == nil || !delivery.IsRedelivery
}
