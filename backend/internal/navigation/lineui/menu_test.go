package lineui

import (
	"bytes"
	"encoding/json"
	"testing"
)

func TestQuickMenuPayload(t *testing.T) {
	for _, tc := range []struct {
		group    bool
		commands []string
	}{
		{false, []string{"出勤紀錄"}},
		{true, []string{"打卡", "出勤", "請假", "漏下班", "摘要"}},
	} {
		data, err := json.Marshal(QuickMenu(tc.group, "https://liff.line.me/example"))
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
		if payload.Type != "text" || len(payload.QuickReply.Items) != len(tc.commands)+1 {
			t.Fatalf("invalid quick reply: %s", data)
		}
		for i, command := range tc.commands {
			item := payload.QuickReply.Items[i]
			if item.Type != "action" || item.Action.Type != "message" || item.Action.Text != command || item.Action.Label != command {
				t.Fatalf("invalid command %q: %s", command, data)
			}
		}
		last := payload.QuickReply.Items[len(payload.QuickReply.Items)-1]
		if last.Action.Type != "uri" || last.Action.Uri != "https://liff.line.me/example" {
			t.Fatalf("invalid URI: %s", data)
		}
	}
}
