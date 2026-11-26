package navigation

import (
	"bytes"
	"encoding/json"
	"strings"
	"testing"

	"github.com/line/line-bot-sdk-go/v8/linebot/messaging_api"
)

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
	message := MiniApp(mobile, desktop).(*messaging_api.TemplateMessage)
	button := message.Template.(*messaging_api.ButtonsTemplate).Actions[0].(*messaging_api.UriAction)
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

func TestMiniAppGroupRepositoryOnBothEntrances(t *testing.T) {
	const id = "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"
	const mobile = "https://miniapp.line.me/example"
	const desktop = "https://app.example.com/miniapp"
	message := MiniApp(mobile, desktop, id).(*messaging_api.TemplateMessage)
	action := message.Template.(*messaging_api.ButtonsTemplate).Actions[0].(*messaging_api.UriAction)
	const query = "?repositories=1&repository=" + id
	if action.Uri != mobile+query || action.AltUri == nil || action.AltUri.Desktop != desktop+query {
		t.Fatalf("group repository was lost between entry points: %+v", action)
	}
}

func TestJournalURLDoesNotGuessRepository(t *testing.T) {
	for _, ids := range [][]string{nil, {"invalid"}, {"aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa", "bbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb"}} {
		if got := JournalURL("https://liff.line.me/example", ids...); got != "https://liff.line.me/example?journal=1&journalCompose=1" {
			t.Fatalf("unexpected journal URL: %s", got)
		}
	}
}

func TestChatRichMenuArtworkLayoutAndActions(t *testing.T) {
	aliases := ChatRichMenuAliases()
	menus := ChatRichMenus("https://miniapp.line.me/example")
	if len(aliases) != 4 || aliases[0] != "rich-menu-functions" || len(menus) != len(aliases) {
		t.Fatalf("invalid private menu set: %v, %v", aliases, menus)
	}
	if menus[aliases[0]].ChatBarText != "收合／展開" {
		t.Fatalf("chat bar must describe its client-owned control: %q", menus[aliases[0]].ChatBarText)
	}
	for i, alias := range aliases[1:] {
		menu := menus[alias]
		if menu.Size.Width != 2500 || menu.Size.Height != 1686 || len(menu.Areas) != 5 {
			t.Fatalf("invalid full-height layout for %s: %+v", alias, menu)
		}
		targets := map[string]bool{}
		for _, area := range menu.Areas[:2] {
			action, ok := area.Action.(*messaging_api.RichMenuSwitchAction)
			if !ok || action.RichMenuAliasId == alias || targets[action.RichMenuAliasId] || area.Bounds.Y != 0 || area.Bounds.Height != 360 {
				t.Fatalf("invalid attendance tab for %s: %+v", alias, area)
			}
			targets[action.RichMenuAliasId] = true
		}
		main, ok := menu.Areas[2].Action.(*messaging_api.PostbackAction)
		if !ok || main.Data != []string{"action=clock-in", "action=support", "action=clock-out"}[i] || menu.Areas[2].Bounds.Y != 360 || menu.Areas[2].Bounds.Height != 966 {
			t.Fatalf("invalid main action for %s: %+v", alias, menu.Areas[2])
		}
		private, ok := menu.Areas[3].Action.(*messaging_api.RichMenuSwitchAction)
		if !ok || private.RichMenuAliasId != "rich-menu-functions" || menu.Areas[3].Bounds.Y != 1326 || menu.Areas[3].Bounds.Width != 1250 {
			t.Fatalf("invalid private menu exit: %+v", menu.Areas[3])
		}
		app, ok := menu.Areas[4].Action.(*messaging_api.UriAction)
		if !ok || app.Uri != "https://miniapp.line.me/example" || menu.Areas[4].Bounds.X != 1250 || menu.Areas[4].Bounds.Y != 1326 {
			t.Fatalf("invalid MINI App exit: %+v", menu.Areas[4])
		}
	}
	if ChatRichMenus("") != nil {
		t.Fatal("invalid MINI App URL must not publish attendance menus")
	}
}

func TestFunctionRichMenuRoutesAndBounds(t *testing.T) {
	const app = "https://miniapp.line.me/example"
	menu := ChatRichMenus(app)["rich-menu-functions"]
	pages := []string{"missing-punch", "correction", "leave", "review", "repositories"}
	labels := []string{"補卡", "更正", "請假", "審核", "儲存庫"}
	if menu == nil || len(menu.Areas) != 9 {
		t.Fatalf("invalid function menu: %+v", menu)
	}
	for _, area := range menu.Areas[:3] {
		if _, ok := area.Action.(*messaging_api.RichMenuSwitchAction); !ok || area.Bounds.Y != 0 || area.Bounds.Height != 360 {
			t.Fatalf("invalid attendance tab: %+v", area)
		}
	}
	for i, page := range pages {
		area := menu.Areas[i+3]
		action, ok := area.Action.(*messaging_api.UriAction)
		if !ok || action.Uri != app+"?menu="+page || action.Label != labels[i] {
			t.Fatalf("area %d link: %+v", i, area.Action)
		}
	}
	appAction, ok := menu.Areas[8].Action.(*messaging_api.UriAction)
	if !ok || appAction.Uri != "https://miniapp.line.me/example" || menu.Areas[8].Bounds.X != 1250 || menu.Areas[8].Bounds.Y != 1326 {
		t.Fatalf("invalid MINI App action: %+v", menu.Areas[8])
	}
}
