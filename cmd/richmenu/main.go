// richmenu previews a three-area menu; external writes require -apply.
package main

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"image"
	_ "image/jpeg"
	_ "image/png"
	"io"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"

	"github.com/963sup/line-mini-app-bot/internal/lineui"
	"github.com/line/line-bot-sdk-go/v8/linebot/messaging_api"
)

func main() {
	if err := run(os.Args[1:]); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func run(args []string) error {
	flags := flag.NewFlagSet("richmenu", flag.ContinueOnError)
	appURL := flags.String("app-url", "", "HTTPS MINI App permanent link")
	imagePath := flags.String("image", "", "2500 x 843 PNG or JPEG, at most 1 MiB")
	apply := flags.Bool("apply", false, "create/upload on LINE; otherwise preview only")
	attendance := flags.Bool("attendance-preview", false, "preview attendance tabs and alias names")
	setDefault := flags.Bool("set-default", false, "also replace the default rich menu when applying")
	menuID := flags.String("menu-id", "", "resume upload using a previously created menu ID")
	if err := flags.Parse(args); err != nil {
		return err
	}
	if flags.NArg() != 0 {
		return errors.New("unexpected positional arguments")
	}
	parsed, err := url.Parse(*appURL)
	if err != nil || parsed.Scheme != "https" || parsed.Hostname() == "" {
		return errors.New("-app-url must be an absolute HTTPS URL")
	}
	if parsed.User != nil || parsed.Fragment != "" {
		return errors.New("-app-url must not contain credentials or a fragment")
	}
	menu := lineui.RichMenu(parsed.String())
	if *attendance {
		if *apply {
			return errors.New("attendance prototype supports preview only")
		}
		return json.NewEncoder(os.Stdout).Encode(lineui.AttendanceRichMenus(parsed.String()))
	}
	if !*apply {
		return json.NewEncoder(os.Stdout).Encode(struct {
			Menu       *messaging_api.RichMenuRequest `json:"menu"`
			MenuID     string                         `json:"existingMenuId,omitempty"`
			SetDefault bool                           `json:"setDefault"`
		}{Menu: menu, MenuID: *menuID, SetDefault: *setDefault})
	}
	content, contentType, err := readMenuImage(*imagePath)
	if err != nil {
		return err
	}
	token := os.Getenv("LINE_MESSAGING_API_CHANNEL_ACCESS_TOKEN")
	if strings.TrimSpace(token) == "" {
		return errors.New("missing LINE_MESSAGING_API_CHANNEL_ACCESS_TOKEN")
	}
	client := &http.Client{Timeout: 10 * time.Second}
	api, err := messaging_api.NewMessagingApiAPI(token, messaging_api.WithHTTPClient(client))
	if err != nil {
		return errors.New("cannot initialize LINE client")
	}
	blob, err := messaging_api.NewMessagingApiBlobAPI(token, messaging_api.WithBlobHTTPClient(client))
	if err != nil {
		return errors.New("cannot initialize LINE image client")
	}
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	api.WithContext(ctx)
	blob.WithContext(ctx)
	if *menuID == "" {
		created, err := api.CreateRichMenu(menu)
		if err != nil {
			return errors.New("rich menu creation failed; inspect LINE configuration and permissions before retrying")
		}
		if created == nil || created.RichMenuId == "" {
			return errors.New("LINE returned no menu ID; inspect existing menus before retrying")
		}
		*menuID = created.RichMenuId
	}
	if _, err := fmt.Fprintln(os.Stdout, "richMenuId:", *menuID); err != nil {
		return errors.New("cannot print rich menu ID; inspect existing menus before retrying")
	}
	if _, err := blob.SetRichMenuImage(*menuID, contentType, bytes.NewReader(content)); err != nil {
		return errors.New("image upload failed; resume with -menu-id using the printed ID; default menu was not changed")
	}
	if *setDefault {
		if _, err := api.SetDefaultRichMenu(*menuID); err != nil {
			return errors.New("setting default failed; inspect the current default before retrying with -menu-id")
		}
		if _, err := fmt.Fprintln(os.Stdout, "default rich menu updated"); err != nil {
			return errors.New("default menu updated but confirmation output failed")
		}
	}
	return nil
}

func readMenuImage(path string) ([]byte, string, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, "", errors.New("provide a readable -image file")
	}
	// This file is read-only; no writes need to be committed on close.
	defer func() { _ = file.Close() }()
	content, err := io.ReadAll(io.LimitReader(file, (1<<20)+1))
	if err != nil || len(content) > 1<<20 {
		return nil, "", errors.New("image is unreadable or exceeds 1 MiB")
	}
	config, format, err := image.DecodeConfig(bytes.NewReader(content))
	if err != nil || (format != "png" && format != "jpeg") {
		return nil, "", errors.New("image must be a valid PNG or JPEG")
	}
	if config.Width != 2500 || config.Height != 843 {
		return nil, "", errors.New("image must be exactly 2500 x 843 pixels")
	}
	return content, "image/" + format, nil
}
