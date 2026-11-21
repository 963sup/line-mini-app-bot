// Package richmenu synchronizes LINE menus supplied by the navigation owner.
package richmenu

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/json"
	"errors"
	"fmt"
	"image"
	_ "image/jpeg"
	_ "image/png"
	"log/slog"
	"net/http"
	"regexp"
	"time"

	"github.com/line/line-bot-sdk-go/v8/linebot/messaging_api"
)

type Menu struct {
	Alias   string
	Request *messaging_api.RichMenuRequest
	Image   []byte
}

var (
	aliasName       = regexp.MustCompile(`^[a-z0-9][a-z0-9-]{0,31}$`)
	managedMenuName = regexp.MustCompile(`^managed-([a-z0-9-]+)-[0-9a-f]{64}$`)
)

// Sync stages every image before switching aliases. The caller must serialize
// publications for the channel; aliases are separate, non-atomic LINE calls.
// menus[0] becomes the default. Clients are copied before setting their context.
func Sync(ctx context.Context, api *messaging_api.MessagingApiAPI, blob *messaging_api.MessagingApiBlobAPI, menus []Menu) error {
	if api == nil || blob == nil || len(menus) == 0 {
		return errors.New("rich menu clients and menus are required")
	}
	client, imagesClient := *api, *blob
	client.WithContext(ctx)
	imagesClient.WithContext(ctx)
	api, blob = &client, &imagesClient
	desired := make([]Menu, len(menus))
	seen := map[string]bool{}
	for i, menu := range menus {
		if !aliasName.MatchString(menu.Alias) || seen[menu.Alias] || menu.Request == nil {
			return errors.New("invalid or duplicate rich menu alias")
		}
		seen[menu.Alias] = true
		if len(menu.Image) == 0 || len(menu.Image) > 1<<20 {
			return errors.New("rich menu image must be at most 1 MiB")
		}
		config, format, err := image.DecodeConfig(bytes.NewReader(menu.Image))
		if err != nil || (format != "jpeg" && format != "png") || config.Width != 2500 || config.Height != 843 {
			return errors.New("rich menu image must be a 2500 x 843 PNG or JPEG")
		}
		data, err := json.Marshal(menu.Request)
		if err != nil {
			return err
		}
		h := sha256.New()
		h.Write(data)
		h.Write(menu.Image)
		request := *menu.Request
		request.Name = fmt.Sprintf("managed-%s-%x", menu.Alias, h.Sum(nil))
		desired[i] = Menu{Alias: menu.Alias, Request: &request, Image: menu.Image}
	}
	list, err := api.GetRichMenuList()
	if err != nil {
		return err
	}
	aliases, err := api.GetRichMenuAliasList()
	if err != nil {
		return err
	}
	ids := map[string]string{}
	for _, menu := range desired {
		hasImage := false
		for _, existing := range list.Richmenus {
			if existing.Name != menu.Request.Name {
				continue
			}
			response, imageResponse, err := blob.GetRichMenuImageWithHttpInfo(existing.RichMenuId)
			if err != nil && (response == nil || response.StatusCode != http.StatusNotFound) {
				return err
			}
			if imageResponse != nil && imageResponse.Body != nil {
				_ = imageResponse.Body.Close()
			}
			ids[menu.Alias] = existing.RichMenuId
			hasImage = err == nil
			break
		}
		if hasImage {
			continue
		}
		if ids[menu.Alias] == "" {
			created, err := api.CreateRichMenu(menu.Request)
			if err != nil {
				return err
			}
			if created == nil || created.RichMenuId == "" {
				return errors.New("LINE returned no rich menu ID")
			}
			ids[menu.Alias] = created.RichMenuId
		}
		_, format, _ := image.DecodeConfig(bytes.NewReader(menu.Image))
		if _, err := blob.SetRichMenuImage(ids[menu.Alias], "image/"+format, bytes.NewReader(menu.Image)); err != nil {
			return errors.New("image upload failed; live aliases unchanged")
		}
	}
	for _, menu := range desired {
		found := false
		for _, existing := range aliases.Aliases {
			if existing.RichMenuAliasId != menu.Alias {
				continue
			}
			found = true
			if existing.RichMenuId != ids[menu.Alias] {
				if _, err := api.UpdateRichMenuAlias(menu.Alias, &messaging_api.UpdateRichMenuAliasRequest{RichMenuId: ids[menu.Alias]}); err != nil {
					return errors.New("alias update failed; next publication can converge")
				}
			}
		}
		if !found {
			if _, err := api.CreateRichMenuAlias(&messaging_api.CreateRichMenuAliasRequest{RichMenuAliasId: menu.Alias, RichMenuId: ids[menu.Alias]}); err != nil {
				return err
			}
		}
	}
	defaultID := ids[desired[0].Alias]
	response, current, err := api.GetDefaultRichMenuIdWithHttpInfo()
	if err != nil && (response == nil || response.StatusCode != http.StatusNotFound) {
		return err
	}
	if current == nil || current.RichMenuId != defaultID {
		if _, err := api.SetDefaultRichMenu(defaultID); err != nil {
			return err
		}
	}
	for _, menu := range desired {
		actual, err := api.GetRichMenuAlias(menu.Alias)
		if err != nil || actual.RichMenuId != ids[menu.Alias] {
			return errors.New("alias readback mismatch")
		}
	}
	actual, err := api.GetDefaultRichMenuId()
	if err != nil || actual.RichMenuId != defaultID {
		return errors.New("default readback mismatch")
	}
	return cleanupMenus(ctx, api, ids)
}

func cleanupMenus(ctx context.Context, api *messaging_api.MessagingApiAPI, current map[string]string) error {
	list, err := api.GetRichMenuList()
	if err != nil {
		return err
	}
	aliases, err := api.GetRichMenuAliasList()
	if err != nil {
		return err
	}
	for _, menu := range list.Richmenus {
		match := managedMenuName.FindStringSubmatch(menu.Name)
		if match == nil {
			continue
		}
		target, managed := current[match[1]]
		if !managed || menu.RichMenuId == target {
			continue
		}
		isCurrent := false
		for _, id := range current {
			if id == menu.RichMenuId {
				isCurrent = true
			}
		}
		if isCurrent {
			continue
		}
		for _, alias := range aliases.Aliases {
			if alias.RichMenuId == menu.RichMenuId {
				return errors.New("cleanup blocked: another alias still references an old managed menu")
			}
		}
		// Rich-menu switching can leave per-user links; migrate all of them first.
		response, _, err := api.RichMenuBatchWithHttpInfo(&messaging_api.RichMenuBatchRequest{Operations: []messaging_api.RichMenuBatchOperationInterface{&messaging_api.RichMenuBatchLinkOperation{From: menu.RichMenuId, To: target}}})
		if err != nil {
			return err
		}
		requestID := response.Header.Get("x-line-request-id")
		if requestID == "" {
			return errors.New("migration returned no request ID; old menu retained")
		}
		deadline := time.Now().Add(45 * time.Second)
		for {
			progress, err := api.GetRichMenuBatchProgress(requestID)
			if err != nil {
				return err
			}
			if progress.Phase == "succeeded" {
				break
			}
			if progress.Phase == "failed" || time.Now().After(deadline) {
				return errors.New("user migration incomplete; old menu retained; rerun sync")
			}
			select {
			case <-ctx.Done():
				return ctx.Err()
			case <-time.After(time.Second):
			}
		}
		if _, err := api.DeleteRichMenu(menu.RichMenuId); err != nil {
			return errors.New("publication verified; old menu deletion failed; rerun sync")
		}
		slog.Info("removed superseded managed menu", "id", menu.RichMenuId)
	}
	return nil
}
