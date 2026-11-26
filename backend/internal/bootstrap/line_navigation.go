package bootstrap

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	firestoreplatform "github.com/963sup/line-mini-app-bot/internal/platform/firestore"
	lineplatform "github.com/963sup/line-mini-app-bot/internal/platform/line"
	"github.com/963sup/line-mini-app-bot/internal/platform/line/richmenu"
	"github.com/963sup/line-mini-app-bot/internal/transport/line/navigation"
	"github.com/line/line-bot-sdk-go/v8/linebot/messaging_api"
)

// SyncLINENavigation runs once in the deployment job, never on service startup.
func SyncLINENavigation(parent context.Context) (result error) {
	ctx, cancel := context.WithTimeout(parent, 8*time.Minute)
	defer cancel()
	auth, _, err := newMiniAppAuthenticator(os.Getenv("LINE_MINI_APP_ENV"))
	if err != nil {
		return err
	}
	config := messagingEnvironment()
	token, channel, project := config.token, config.channel, os.Getenv("GOOGLE_CLOUD_PROJECT")
	if token == "" || !validChannelID(channel) || project == "" {
		return errors.New("LINE navigation job requires Messaging API token, channel ID and GOOGLE_CLOUD_PROJECT")
	}
	dir := os.Getenv("LINE_RICH_MENU_IMAGES_DIR")
	if dir == "" {
		dir = "/assets/line/richmenu"
	}
	requests := navigation.ChatRichMenus(lineplatform.MiniAppURL(auth.LIFFID))
	if requests == nil {
		return errors.New("LINE navigation job requires a valid MINI App URL")
	}
	menus := make([]richmenu.Menu, 0, len(requests))
	for _, alias := range navigation.ChatRichMenuAliases() {
		suffix, ok := strings.CutPrefix(alias, "rich-menu-")
		if !ok {
			return fmt.Errorf("invalid chat rich menu alias %q", alias)
		}
		filename := "rich_menu_" + strings.ReplaceAll(suffix, "-", "_") + ".png"
		image, err := os.ReadFile(filepath.Join(dir, filename))
		if err != nil {
			return fmt.Errorf("read rich menu image %s: %w", alias, err)
		}
		menus = append(menus, richmenu.Menu{Alias: alias, Request: requests[alias], Image: image})
	}
	database := os.Getenv("FIRESTORE_DATABASE_ID")
	db, err := firestoreplatform.NewClient(ctx, project, database)
	if err != nil {
		return err
	}
	defer func() { result = errors.Join(result, db.Close()) }()
	// The lease outlives the entire bounded job by two minutes. Every LINE call
	// shares ctx, so the holder stops before another execution can reclaim it.
	release, err := richmenu.AcquireLease(ctx, db, channel)
	if err != nil {
		return err
	}
	defer func() { result = errors.Join(result, release()) }()
	client := &http.Client{Timeout: 20 * time.Second, CheckRedirect: func(*http.Request, []*http.Request) error { return http.ErrUseLastResponse }}
	api, err := messaging_api.NewMessagingApiAPI(token, messaging_api.WithHTTPClient(client))
	if err != nil {
		return err
	}
	blob, err := messaging_api.NewMessagingApiBlobAPI(token, messaging_api.WithBlobHTTPClient(client))
	if err != nil {
		return err
	}
	return richmenu.Sync(ctx, api, blob, menus, map[string]string{
		"attendance-in":     "rich-menu-clock-in",
		"attendance-out":    "rich-menu-clock-out",
		"rich-menu-private": "rich-menu-functions",
	})
}
