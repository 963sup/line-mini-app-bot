package main

import (
	"bytes"
	"crypto/sha256"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"github.com/963sup/line-mini-app-bot/internal/navigation/lineui"
	"github.com/line/line-bot-sdk-go/v8/linebot/messaging_api"
)

// syncMenus stages both immutable menus before changing either live alias.
// LINE itself retains the content-addressed publication state across CI runs.
func syncMenus(args []string) error {
	f := flag.NewFlagSet("sync", flag.ContinueOnError)
	app := f.String("app-url", "", "HTTPS MINI App URL")
	dir := f.String("images", "assets/richmenu", "directory containing the two JPEG images")
	apply := f.Bool("apply", false, "publish changes on LINE")
	if err := f.Parse(args); err != nil {
		return err
	}
	u, err := url.Parse(*app)
	if err != nil || u.Scheme != "https" || u.Hostname() == "" || u.User != nil || u.Fragment != "" || f.NArg() != 0 {
		return errors.New("provide -app-url with an absolute HTTPS URL")
	}
	names := []string{"attendance-in", "attendance-out"}
	menus := lineui.AttendanceRichMenus(*app)
	images := map[string][]byte{}
	for _, alias := range names {
		data, _, err := readMenuImage(filepath.Join(*dir, alias+".jpg"))
		if err != nil {
			return fmt.Errorf("%s: %w", alias, err)
		}
		config, err := json.Marshal(menus[alias])
		if err != nil {
			return err
		}
		h := sha256.New()
		h.Write(config)
		h.Write(data)
		menus[alias].Name = fmt.Sprintf("managed-%s-%x", alias, h.Sum(nil))
		images[alias] = data
		fmt.Println(alias, menus[alias].Name)
	}
	if !*apply {
		return nil
	}
	token := os.Getenv("LINE_MESSAGING_API_CHANNEL_ACCESS_TOKEN")
	if token == "" {
		return errors.New("missing LINE_MESSAGING_API_CHANNEL_ACCESS_TOKEN")
	}
	httpClient := &http.Client{Timeout: 20 * time.Second}
	api, err := messaging_api.NewMessagingApiAPI(token, messaging_api.WithHTTPClient(httpClient))
	if err != nil {
		return errors.New("cannot initialize LINE client")
	}
	blob, err := messaging_api.NewMessagingApiBlobAPI(token, messaging_api.WithBlobHTTPClient(httpClient))
	if err != nil {
		return errors.New("cannot initialize image client")
	}
	list, err := api.GetRichMenuList()
	if err != nil {
		return errors.New("cannot list LINE menus")
	}
	aliases, err := api.GetRichMenuAliasList()
	if err != nil {
		return errors.New("cannot list LINE aliases")
	}
	ids := map[string]string{}
	for _, alias := range names {
		for _, menu := range list.Richmenus {
			if menu.Name == menus[alias].Name {
				// Creation may have succeeded while image upload failed on a previous run.
				if _, err := blob.GetRichMenuImage(menu.RichMenuId); err == nil {
					ids[alias] = menu.RichMenuId
					break
				}
			}
		}
		if ids[alias] != "" {
			continue
		}
		created, err := api.CreateRichMenu(menus[alias])
		if err != nil {
			return errors.New("menu creation failed; live aliases unchanged")
		}
		ids[alias] = created.RichMenuId
		if _, err := blob.SetRichMenuImage(created.RichMenuId, "image/jpeg", bytes.NewReader(images[alias])); err != nil {
			return errors.New("image upload failed; live aliases unchanged; rerun sync")
		}
	}
	// Alias updates are individual API operations, not an atomic LINE transaction.
	// Both target images exist now; a retry converges after a partial update.
	for _, alias := range names {
		found := false
		for _, existing := range aliases.Aliases {
			if existing.RichMenuAliasId != alias {
				continue
			}
			found = true
			if existing.RichMenuId != ids[alias] {
				if _, err := api.UpdateRichMenuAlias(alias, &messaging_api.UpdateRichMenuAliasRequest{RichMenuId: ids[alias]}); err != nil {
					return errors.New("alias update failed; rerun sync to converge")
				}
			}
		}
		if !found {
			if _, err := api.CreateRichMenuAlias(&messaging_api.CreateRichMenuAliasRequest{RichMenuAliasId: alias, RichMenuId: ids[alias]}); err != nil {
				return errors.New("alias creation failed; rerun sync")
			}
		}
	}
	current, err := api.GetDefaultRichMenuId()
	if err != nil || current.RichMenuId != ids[names[0]] {
		if _, err := api.SetDefaultRichMenu(ids[names[0]]); err != nil {
			return errors.New("default update failed; rerun sync")
		}
	}
	for _, alias := range names {
		actual, err := api.GetRichMenuAlias(alias)
		if err != nil || actual.RichMenuId != ids[alias] {
			return errors.New("alias readback mismatch")
		}
	}
	actual, err := api.GetDefaultRichMenuId()
	if err != nil || actual.RichMenuId != ids[names[0]] {
		return errors.New("default readback mismatch")
	}
	if err := cleanupMenus(api, ids); err != nil {
		return err
	}
	fmt.Println("Rich Menu sync verified")
	return nil
}

var managedMenuName = regexp.MustCompile(`^managed-(attendance-in|attendance-out)-[0-9a-f]{64}$`)

func cleanupMenus(api *messaging_api.MessagingApiAPI, current map[string]string) error {
	list, err := api.GetRichMenuList()
	if err != nil {
		return errors.New("publication verified; cannot list cleanup candidates; rerun sync")
	}
	aliases, err := api.GetRichMenuAliasList()
	if err != nil {
		return errors.New("publication verified; cannot check cleanup alias references")
	}
	for _, menu := range list.Richmenus {
		if !managedMenuName.MatchString(menu.Name) || menu.RichMenuId == current["attendance-in"] || menu.RichMenuId == current["attendance-out"] {
			continue
		}
		for _, alias := range aliases.Aliases {
			if alias.RichMenuId == menu.RichMenuId {
				return errors.New("cleanup blocked: another alias still references an old managed menu")
			}
		}
		target := current["attendance-in"]
		if strings.HasPrefix(menu.Name, "managed-attendance-out-") {
			target = current["attendance-out"]
		}
		// Rich-menu switching can leave per-user links; migrate all of them first.
		response, _, err := api.RichMenuBatchWithHttpInfo(&messaging_api.RichMenuBatchRequest{Operations: []messaging_api.RichMenuBatchOperationInterface{&messaging_api.RichMenuBatchLinkOperation{From: menu.RichMenuId, To: target}}})
		if err != nil {
			return errors.New("publication verified; user menu migration failed; old menu retained")
		}
		requestID := response.Header.Get("x-line-request-id")
		if requestID == "" {
			return errors.New("migration returned no request ID; old menu retained")
		}
		deadline := time.Now().Add(45 * time.Second)
		for {
			progress, err := api.GetRichMenuBatchProgress(requestID)
			if err != nil {
				return errors.New("cannot verify user migration; old menu retained; rerun sync")
			}
			if progress.Phase == "succeeded" {
				break
			}
			if progress.Phase == "failed" || time.Now().After(deadline) {
				return errors.New("user migration incomplete; old menu retained; rerun sync")
			}
			time.Sleep(time.Second)
		}
		if _, err := api.DeleteRichMenu(menu.RichMenuId); err != nil {
			return errors.New("publication verified; old menu deletion failed; rerun sync")
		}
		fmt.Println("removed superseded managed menu:", menu.RichMenuId)
	}
	return nil
}
