package bootstrap

import (
	"context"
	"crypto/rand"
	"errors"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"cloud.google.com/go/firestore"
	"github.com/963sup/line-mini-app-bot/internal/platform/line/richmenu"
	"github.com/963sup/line-mini-app-bot/internal/transport/line/navigation"
	"github.com/line/line-bot-sdk-go/v8/linebot/messaging_api"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// SyncLINENavigation runs once in the deployment job, never on service startup.
func SyncLINENavigation(parent context.Context) (result error) {
	ctx, cancel := context.WithTimeout(parent, 8*time.Minute)
	defer cancel()
	auth, _, err := newMiniAppAuthenticator(os.Getenv("LINE_MINI_APP_ENV"))
	if err != nil {
		return err
	}
	token, channel, project := os.Getenv("LINE_MESSAGING_API_CHANNEL_ACCESS_TOKEN"), os.Getenv("LINE_MESSAGING_API_CHANNEL_ID"), os.Getenv("GOOGLE_CLOUD_PROJECT")
	if token == "" || channel == "" || strings.Trim(channel, "0123456789") != "" || project == "" {
		return errors.New("LINE navigation job requires Messaging API token, channel ID and GOOGLE_CLOUD_PROJECT")
	}
	dir := os.Getenv("LINE_RICH_MENU_IMAGES_DIR")
	if dir == "" {
		dir = "/assets/line/richmenu"
	}
	definitions := navigation.AttendanceRichMenus("https://miniapp.line.me/" + auth.LIFFID)
	menus := make([]richmenu.Menu, 0, len(definitions))
	for _, alias := range []string{"attendance-in", "attendance-out"} {
		image, err := os.ReadFile(filepath.Join(dir, alias+".jpg"))
		if err != nil {
			return fmt.Errorf("read rich menu image %s: %w", alias, err)
		}
		menus = append(menus, richmenu.Menu{Alias: alias, Request: definitions[alias], Image: image})
	}
	database := os.Getenv("FIRESTORE_DATABASE_ID")
	if database == "" {
		database = "(default)"
	}
	db, err := firestore.NewClientWithDatabase(ctx, project, database)
	if err != nil {
		return err
	}
	defer func() { result = errors.Join(result, db.Close()) }()
	// The lease outlives the entire bounded job by two minutes. Every LINE call
	// shares ctx, so the holder stops before another execution can reclaim it.
	release, err := acquireNavigationLease(ctx, db, channel)
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
	return richmenu.Sync(ctx, api, blob, menus)
}

var errNavigationBusy = errors.New("another LINE navigation publication is active")

func acquireNavigationLease(ctx context.Context, db *firestore.Client, channel string) (func() error, error) {
	ref := db.Collection("line_navigation_publication_leases").Doc(channel)
	owner := rand.Text()
	for {
		err := db.RunTransaction(ctx, func(ctx context.Context, tx *firestore.Transaction) error {
			doc, err := tx.Get(ref)
			if err != nil && status.Code(err) != codes.NotFound {
				return err
			}
			if err == nil {
				until, ok := doc.Data()["expiresAt"].(time.Time)
				if !ok || until.IsZero() {
					return errors.New("invalid LINE navigation lease; refusing publication")
				}
				if until.After(time.Now()) {
					return errNavigationBusy
				}
			}
			return tx.Set(ref, map[string]any{"owner": owner, "expiresAt": time.Now().Add(10 * time.Minute)})
		})
		if err == nil {
			break
		}
		if !errors.Is(err, errNavigationBusy) {
			return nil, err
		}
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		case <-time.After(5 * time.Second):
		}
	}
	return func() error {
		releaseCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		return db.RunTransaction(releaseCtx, func(ctx context.Context, tx *firestore.Transaction) error {
			doc, err := tx.Get(ref)
			if status.Code(err) == codes.NotFound {
				return nil
			}
			if err != nil {
				return err
			}
			if doc.Data()["owner"] != owner {
				return nil
			}
			return tx.Delete(ref)
		})
	}, nil
}
