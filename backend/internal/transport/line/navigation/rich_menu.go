package navigation

import (
	_ "embed"
	"encoding/json"

	"github.com/line/line-bot-sdk-go/v8/linebot/messaging_api"
)

//go:embed rich_menu.json
var attendanceMenuJSON []byte

var attendanceMenu = loadAttendanceMenu()

type menuPage struct {
	Alias, Label, Postback, Hint string
}

type menuBounds struct {
	X, Y, Width, Height int64
}

type menuFeature struct {
	Label, Feature string
	Bounds         menuBounds
}

type menuLayout struct {
	Width, Height, TabHeight, FooterHeight int64
	TabWidths                              []int64
	ChatBarText, MiniAppLabel              string
	FunctionPage                           menuPage
	Features                               []menuFeature
	Pages                                  []menuPage
}

func loadAttendanceMenu() menuLayout {
	var layout menuLayout
	if err := json.Unmarshal(attendanceMenuJSON, &layout); err != nil {
		panic(err)
	}
	if layout.Width != 2500 || layout.Height != 1686 || layout.TabHeight <= 0 || layout.FooterHeight <= 0 ||
		layout.TabHeight+layout.FooterHeight >= layout.Height || layout.MiniAppLabel == "" ||
		layout.FunctionPage.Alias == "" || layout.FunctionPage.Label == "" || layout.FunctionPage.Hint == "" || len(layout.Features) == 0 ||
		len(layout.Pages) != 3 || len(layout.TabWidths) != len(layout.Pages) {
		panic("invalid attendance rich menu layout")
	}
	var width int64
	aliases := map[string]bool{}
	for i, page := range layout.Pages {
		if layout.TabWidths[i] <= 0 || page.Alias == "" || aliases[page.Alias] || page.Label == "" || page.Postback == "" || page.Hint == "" {
			panic("invalid attendance rich menu page")
		}
		width += layout.TabWidths[i]
		aliases[page.Alias] = true
	}
	if width != layout.Width {
		panic("invalid attendance rich menu tab widths")
	}
	if aliases[layout.FunctionPage.Alias] {
		panic("duplicate function rich menu alias")
	}
	for _, feature := range layout.Features {
		bounds := feature.Bounds
		if feature.Label == "" || feature.Feature == "" || bounds.Width <= 0 || bounds.Height <= 0 || bounds.X < 0 ||
			bounds.Y < layout.TabHeight || bounds.X+bounds.Width > layout.Width || bounds.Y+bounds.Height > layout.Height-layout.FooterHeight {
			panic("invalid function rich menu feature")
		}
	}
	return layout
}

// ChatRichMenuAliases lists the pages in publication order, with the function page first as the default.
func ChatRichMenuAliases() []string {
	aliases := make([]string, 0, len(attendanceMenu.Pages)+1)
	aliases = append(aliases, attendanceMenu.FunctionPage.Alias)
	for _, page := range attendanceMenu.Pages {
		aliases = append(aliases, page.Alias)
	}
	return aliases
}

func ChatRichMenus(appURL string) map[string]*messaging_api.RichMenuRequest {
	layout := attendanceMenu
	miniAppURL := featureURL(appURL, "", "")
	if miniAppURL == "" {
		return nil
	}
	menus := make(map[string]*messaging_api.RichMenuRequest, len(layout.Pages)+1)
	for i, page := range layout.Pages {
		menu := &messaging_api.RichMenuRequest{
			Size: &messaging_api.RichMenuSize{Width: layout.Width, Height: layout.Height},
			Name: "出勤：" + page.Label, ChatBarText: layout.ChatBarText,
		}
		var x int64
		for j, target := range layout.Pages {
			width := layout.TabWidths[j]
			if j != i {
				menu.Areas = append(menu.Areas, messaging_api.RichMenuArea{
					Bounds: &messaging_api.RichMenuBounds{X: x, Width: width, Height: layout.TabHeight},
					Action: &messaging_api.RichMenuSwitchAction{Label: target.Label, RichMenuAliasId: target.Alias, Data: "tab=" + target.Alias},
				})
			}
			x += width
		}
		footerY := layout.Height - layout.FooterHeight
		menu.Areas = append(menu.Areas,
			messaging_api.RichMenuArea{
				Bounds: &messaging_api.RichMenuBounds{Y: layout.TabHeight, Width: layout.Width, Height: footerY - layout.TabHeight},
				Action: &messaging_api.PostbackAction{Label: page.Label, Data: page.Postback},
			},
			messaging_api.RichMenuArea{
				Bounds: &messaging_api.RichMenuBounds{Y: footerY, Width: layout.Width / 2, Height: layout.FooterHeight},
				Action: &messaging_api.RichMenuSwitchAction{Label: layout.FunctionPage.Label, RichMenuAliasId: layout.FunctionPage.Alias, Data: "tab=" + layout.FunctionPage.Alias},
			},
			messaging_api.RichMenuArea{
				Bounds: &messaging_api.RichMenuBounds{X: layout.Width / 2, Y: footerY, Width: layout.Width / 2, Height: layout.FooterHeight},
				Action: &messaging_api.UriAction{Label: layout.MiniAppLabel, Uri: miniAppURL},
			},
		)
		menus[page.Alias] = menu
	}
	function := &messaging_api.RichMenuRequest{
		Size: &messaging_api.RichMenuSize{Width: layout.Width, Height: layout.Height},
		Name: layout.FunctionPage.Label, ChatBarText: layout.ChatBarText,
	}
	var x int64
	for i, target := range layout.Pages {
		width := layout.TabWidths[i]
		function.Areas = append(function.Areas, messaging_api.RichMenuArea{
			Bounds: &messaging_api.RichMenuBounds{X: x, Width: width, Height: layout.TabHeight},
			Action: &messaging_api.RichMenuSwitchAction{Label: target.Label, RichMenuAliasId: target.Alias, Data: "tab=" + target.Alias},
		})
		x += width
	}
	for _, entry := range layout.Features {
		link := featureURL(appURL, entry.Feature, "")
		if link == "" {
			return nil
		}
		bounds := entry.Bounds
		function.Areas = append(function.Areas, messaging_api.RichMenuArea{
			Bounds: &messaging_api.RichMenuBounds{X: bounds.X, Y: bounds.Y, Width: bounds.Width, Height: bounds.Height},
			Action: &messaging_api.UriAction{Label: entry.Label, Uri: link},
		})
	}
	footerY := layout.Height - layout.FooterHeight
	function.Areas = append(function.Areas, messaging_api.RichMenuArea{
		Bounds: &messaging_api.RichMenuBounds{X: layout.Width / 2, Y: footerY, Width: layout.Width / 2, Height: layout.FooterHeight},
		Action: &messaging_api.UriAction{Label: layout.MiniAppLabel, Uri: miniAppURL},
	})
	menus[layout.FunctionPage.Alias] = function
	return menus
}
