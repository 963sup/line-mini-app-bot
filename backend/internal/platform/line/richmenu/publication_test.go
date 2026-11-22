package richmenu

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"image"
	"image/jpeg"
	"io"
	"net/http"
	"strings"
	"testing"

	"github.com/line/line-bot-sdk-go/v8/linebot/messaging_api"
)

type publication struct {
	menus                 map[string]string
	images                map[string]bool
	aliases               map[string]string
	defaultID             string
	writes, creates       int
	failUpload, failAlias bool
	defaultStatus         int
	migrationPhase        string
	migrations            map[string]string
}

func (p *publication) RoundTrip(r *http.Request) (*http.Response, error) {
	path := strings.TrimPrefix(r.URL.Path, "/v2/bot/")
	code, body := 200, `{}`
	header := make(http.Header)
	if r.Method != "GET" {
		p.writes++
	}
	switch {
	case path == "richmenu/list":
		var menus []map[string]string
		for id, name := range p.menus {
			menus = append(menus, map[string]string{"richMenuId": id, "name": name})
		}
		data, _ := json.Marshal(map[string]any{"richmenus": menus})
		body = string(data)
	case path == "richmenu/alias/list":
		var aliases []map[string]string
		for alias, id := range p.aliases {
			aliases = append(aliases, map[string]string{"richMenuAliasId": alias, "richMenuId": id})
		}
		data, _ := json.Marshal(map[string]any{"aliases": aliases})
		body = string(data)
	case path == "richmenu/batch":
		var batch struct {
			Operations []struct{ From, To string }
		}
		if err := json.NewDecoder(r.Body).Decode(&batch); err != nil || len(batch.Operations) != 1 {
			return nil, fmt.Errorf("invalid migration request")
		}
		op := batch.Operations[0]
		if p.migrations == nil {
			p.migrations = make(map[string]string)
		}
		p.migrations[op.From] = op.To
		header.Set("X-Line-Request-Id", "migration-test")
	case path == "richmenu/progress/batch":
		body = `{"phase":"` + p.migrationPhase + `"}`
	case r.Method == "DELETE" && strings.HasPrefix(path, "richmenu/"):
		id := strings.TrimPrefix(path, "richmenu/")
		if p.migrationPhase != "succeeded" || p.migrations[id] == "" {
			return nil, fmt.Errorf("deleted %s before migration succeeded", id)
		}
		delete(p.menus, id)
		delete(p.images, id)
	case path == "richmenu" && r.Method == "POST":
		var menu struct {
			Name string `json:"name"`
		}
		_ = json.NewDecoder(r.Body).Decode(&menu)
		p.creates++
		id := fmt.Sprint(p.creates)
		p.menus[id] = menu.Name
		body = `{"richMenuId":"` + id + `"}`
	case strings.HasSuffix(path, "/content"):
		id := strings.Split(path, "/")[1]
		if r.Method == "POST" {
			if p.failUpload {
				code = 500
			} else {
				p.images[id] = true
			}
		} else if !p.images[id] {
			code = 404
		}
	case path == "richmenu/alias" || strings.HasPrefix(path, "richmenu/alias/"):
		alias := strings.TrimPrefix(path, "richmenu/alias/")
		if r.Method == "GET" {
			body = `{"richMenuAliasId":"` + alias + `","richMenuId":"` + p.aliases[alias] + `"}`
			break
		}
		var value struct {
			Alias string `json:"richMenuAliasId"`
			ID    string `json:"richMenuId"`
		}
		_ = json.NewDecoder(r.Body).Decode(&value)
		if path == "richmenu/alias" {
			alias = value.Alias
		}
		if p.failAlias && alias == "tab-two" {
			code = 500
			break
		}
		p.aliases[alias] = value.ID
	case path == "user/all/richmenu":
		if p.defaultStatus != 0 {
			code = p.defaultStatus
		} else if p.defaultID == "" {
			code = 404
		} else {
			body = `{"richMenuId":"` + p.defaultID + `"}`
		}
	case strings.HasPrefix(path, "user/all/richmenu/"):
		p.defaultID = strings.TrimPrefix(path, "user/all/richmenu/")
	default:
		return nil, fmt.Errorf("unexpected operation %s %s", r.Method, path)
	}
	if code >= 400 {
		body = `{"message":"failure"}`
	}
	return &http.Response{StatusCode: code, Header: header, Body: io.NopCloser(strings.NewReader(body))}, nil
}

func publicationFixture(t *testing.T) (*publication, *messaging_api.MessagingApiAPI, *messaging_api.MessagingApiBlobAPI, []Menu) {
	t.Helper()
	p := &publication{menus: map[string]string{}, images: map[string]bool{}, aliases: map[string]string{}}
	client := &http.Client{Transport: p}
	api, err := messaging_api.NewMessagingApiAPI("test", messaging_api.WithHTTPClient(client))
	if err != nil {
		t.Fatal(err)
	}
	blob, err := messaging_api.NewMessagingApiBlobAPI("test", messaging_api.WithBlobHTTPClient(client))
	if err != nil {
		t.Fatal(err)
	}
	var imageData bytes.Buffer
	if err := jpeg.Encode(&imageData, image.NewGray(image.Rect(0, 0, 2500, 1686)), nil); err != nil {
		t.Fatal(err)
	}
	menus := []Menu{{Alias: "tab-one", Request: &messaging_api.RichMenuRequest{Name: "One", Size: &messaging_api.RichMenuSize{Width: 2500, Height: 1686}}, Image: imageData.Bytes()}, {Alias: "tab-two", Request: &messaging_api.RichMenuRequest{Name: "Two", Size: &messaging_api.RichMenuSize{Width: 2500, Height: 1686}}, Image: imageData.Bytes()}}
	return p, api, blob, menus
}

func TestSyncStagesAllImagesAndConvergesWithoutFurtherWrites(t *testing.T) {
	p, api, blob, menus := publicationFixture(t)
	p.failUpload = true
	if err := Sync(context.Background(), api, blob, menus); err == nil {
		t.Fatal("upload failure accepted")
	}
	if len(p.aliases) != 0 || p.defaultID != "" {
		t.Fatal("failed staging changed live navigation")
	}
	p.failUpload = false
	if err := Sync(context.Background(), api, blob, menus); err != nil {
		t.Fatal(err)
	}
	writes := p.writes
	if p.defaultID != p.aliases["tab-one"] || !p.images[p.aliases["tab-two"]] {
		t.Fatal("publication incomplete")
	}
	if err := Sync(context.Background(), api, blob, menus); err != nil {
		t.Fatal(err)
	}
	if p.writes != writes || p.creates != 2 {
		t.Fatal("unchanged publication wrote to LINE")
	}
	if menus[0].Request.Name != "One" {
		t.Fatal("Sync mutated navigation configuration")
	}
}

func TestSyncAliasFailureCanConvergeAndDefaultLookupFailureStops(t *testing.T) {
	p, api, blob, menus := publicationFixture(t)
	p.failAlias = true
	if err := Sync(context.Background(), api, blob, menus); err == nil || p.defaultID != "" || len(p.aliases) != 1 {
		t.Fatal("partial alias failure did not preserve the old default")
	}
	p.failAlias = false
	if err := Sync(context.Background(), api, blob, menus); err != nil {
		t.Fatal(err)
	}
	if p.creates != 2 {
		t.Fatal("retry duplicated already staged menus")
	}
	p.defaultStatus = 500
	writes := p.writes
	if err := Sync(context.Background(), api, blob, menus); err == nil {
		t.Fatal("default lookup failure ignored")
	}
	if p.writes != writes {
		t.Fatal("lookup failure caused a default write")
	}
}

func TestSyncRejectsRequestSizeMismatchBeforePublishing(t *testing.T) {
	p, api, blob, menus := publicationFixture(t)
	menus[1].Request.Size.Height = 843
	if err := Sync(context.Background(), api, blob, menus); err == nil {
		t.Fatal("accepted a request whose size differs from its image")
	}
	if p.writes != 0 {
		t.Fatal("size mismatch caused a LINE write")
	}
}

func TestSyncMigratesRetiredAliasesAndRetriesSafely(t *testing.T) {
	p, api, blob, menus := publicationFixture(t)
	p.menus["legacy"] = "managed-attendance-in-" + strings.Repeat("a", 64)
	p.aliases["attendance-in"] = "legacy"
	p.migrationPhase = "failed"
	legacy := map[string]string{"attendance-in": "tab-one", "attendance-out": "tab-two"}
	if err := SyncWithLegacyAliases(context.Background(), api, blob, menus, legacy); err == nil {
		t.Fatal("failed migration was accepted")
	}
	target := p.aliases["tab-one"]
	if target == "" || p.aliases["attendance-in"] != target || p.menus["legacy"] == "" || p.migrations["legacy"] != target {
		t.Fatalf("expected the old menu to survive with its alias redirected: %+v", p)
	}
	if _, exists := p.aliases["attendance-out"]; exists {
		t.Fatal("created an unused retired alias")
	}
	p.migrationPhase = "succeeded"
	if err := SyncWithLegacyAliases(context.Background(), api, blob, menus, legacy); err != nil || p.menus["legacy"] != "" {
		t.Fatalf("retry did not finish the migration: %v", err)
	}
	menus[0].Request.Name = "new revision"
	if err := SyncWithLegacyAliases(context.Background(), api, blob, menus, legacy); err != nil {
		t.Fatalf("later publication failed to keep the retired alias current: %v", err)
	}
	if p.aliases["attendance-in"] != p.aliases["tab-one"] || p.migrations[target] != p.aliases["tab-one"] {
		t.Fatal("compatibility alias or existing users remained on the previous release")
	}
	writes := p.writes
	if err := SyncWithLegacyAliases(context.Background(), api, blob, menus, legacy); err != nil || p.writes != writes {
		t.Fatalf("converged publication wrote again: %v", err)
	}
}

func TestSyncDoesNotRedirectUnmanagedLegacyAlias(t *testing.T) {
	p, api, blob, menus := publicationFixture(t)
	p.menus["foreign"] = "managed-by-another-tool"
	p.aliases["attendance-in"] = "foreign"
	err := SyncWithLegacyAliases(context.Background(), api, blob, menus, map[string]string{"attendance-in": "tab-one"})
	if err == nil || p.aliases["attendance-in"] != "foreign" || p.menus["foreign"] == "" || p.writes != 0 {
		t.Fatal("unmanaged legacy alias was overridden")
	}
}
