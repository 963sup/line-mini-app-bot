package lineui

import (
	"bytes"
	"encoding/json"
	"strings"
	"testing"

	"github.com/line/line-bot-sdk-go/v8/linebot/messaging_api"
)

func TestQuickMenuPayload(t *testing.T) {
	for _, tc := range []struct {
		group    bool
		commands []string
	}{
		{false, []string{"出勤紀錄"}},
		{true, []string{"打卡", "出勤", "請假", "補卡", "摘要"}},
	} {
		data, err := json.Marshal(QuickMenu(tc.group, "https://liff.line.me/example", "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"))
		if err != nil {
			t.Fatal(err)
		}
		var payload struct {
			Type       string `json:"type"`
			QuickReply struct {
				Items []struct {
					Type   string `json:"type"`
					Action struct {
						Type  string `json:"type"`
						Text  string `json:"text"`
						Label string `json:"label"`
						Uri   string `json:"uri"`
						Data  string `json:"data"`
					} `json:"action"`
				} `json:"items"`
			} `json:"quickReply"`
		}
		if err := json.Unmarshal(data, &payload); err != nil {
			t.Fatal(err)
		}
		if bytes.Contains(data, []byte("關閉摘要")) || bytes.Contains(data, []byte("開啟摘要")) {
			t.Fatal("legacy switches must not appear")
		}
		if payload.Type != "text" || len(payload.QuickReply.Items) != len(tc.commands)+2+map[bool]int{true: 1}[tc.group] {
			t.Fatalf("invalid quick reply: %s", data)
		}
		for i, command := range tc.commands {
			index := i
			if tc.group && i > 0 {
				index++
			}
			item := payload.QuickReply.Items[index]
			if item.Type != "action" || item.Action.Type != "message" || item.Action.Text != command || item.Action.Label != command {
				t.Fatalf("invalid command %q: %s", command, data)
			}
		}
		if tc.group {
			journal := payload.QuickReply.Items[1].Action
			if journal.Type != "uri" || journal.Label != "日誌" || journal.Uri != "https://liff.line.me/example?journal=1" {
				t.Fatalf("invalid journal link: %s", data)
			}
		}
		space := payload.QuickReply.Items[len(payload.QuickReply.Items)-2].Action
		expectedSpace := "https://liff.line.me/example?repositories=1"
		if tc.group {
			expectedSpace = "https://liff.line.me/example?repositories=1&repository=aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"
		}
		if space.Type != "uri" || space.Label != "空間" || space.Uri != expectedSpace {
			t.Fatalf("invalid repository link: %s", data)
		}
		last := payload.QuickReply.Items[len(payload.QuickReply.Items)-1]
		if last.Action.Type != "postback" || last.Action.Data != "action=miniapp" || last.Action.Uri != "" {
			t.Fatalf("invalid URI: %s", data)
		}
	}
}

func TestMiniAppDesktopDestinationValidation(t *testing.T) {
	mobile := "https://miniapp.line.me/123-test?repositories=1&repository=aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"
	for _, endpoint := range []string{"", "http://app.example/miniapp", "javascript:alert(1)", "https://user:password@app.example/miniapp", "https://miniapp.line.me/123-test", "https://liff.line.me/123-test"} {
		action := miniAppAction(mobile, endpoint)
		if action.Uri != mobile || action.AltUri != nil {
			t.Fatal("invalid desktop URL must not replace or change mobile URL")
		}
		data, _ := json.Marshal(MiniApp(mobile, endpoint))
		if !bytes.Contains(data, []byte("此入口目前供手機")) {
			t.Fatal("missing/invalid configuration must be explicit")
		}
	}
	desktop := "https://app.example/miniapp?view=desktop"
	action := miniAppAction(mobile, desktop)
	if action.AltUri == nil || action.AltUri.Desktop != desktop || action.Uri != mobile {
		t.Fatal("configured destinations must remain exact")
	}
	help := Help(false, mobile, desktop).(*messaging_api.FlexMessage)
	footer := help.Contents.(*messaging_api.FlexBubble).Footer.Contents
	button := footer[len(footer)-1].(*messaging_api.FlexButton).Action.(*messaging_api.UriAction)
	if button.AltUri == nil || button.AltUri.Desktop != desktop {
		t.Fatal("Help and standalone card must share desktop routing")
	}
	if _, ok := MiniApp("", desktop).(*messaging_api.TextMessage); !ok {
		t.Fatal("missing mobile configuration must not render a broken button")
	}
}

func TestMiniAppDesktopEncodedURLLimit(t *testing.T) {
	prefix := "https://app.example/"
	atLimit := prefix + strings.Repeat("a", 10000-len(prefix))
	if miniAppAction("https://miniapp.line.me/123-test", atLimit).AltUri == nil {
		t.Fatal("documented 10000-character endpoint should be accepted")
	}
	if miniAppAction("https://miniapp.line.me/123-test", atLimit+"a").AltUri != nil {
		t.Fatal("endpoint beyond official bound must use fallback")
	}
	if miniAppAction("https://miniapp.line.me/123-test", prefix+strings.Repeat("中", 2000)).AltUri != nil {
		t.Fatal("bound applies after UTF-8 percent encoding")
	}
}
