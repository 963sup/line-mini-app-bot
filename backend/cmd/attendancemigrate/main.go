// attendancemigrate prepares query metadata. Dry-run is the default; it never deploys indexes.
package main

import (
	"context"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"path/filepath"
	"strings"
	"syscall"

	db "cloud.google.com/go/firestore"
	storage "github.com/963sup/line-mini-app-bot/internal/attendance/firestore"
)

type checkpoint struct {
	Project  string                 `json:"project"`
	Database string                 `json:"database"`
	Progress storage.QueryMigration `json:"progress"`
}

func main() {
	if err := run(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func run() (resultErr error) {
	apply := flag.Bool("apply", false, "write derived query metadata; does not switch application reads")
	activate := flag.Bool("activate", false, "activate indexed reads after completed apply; requires indexes and all writers ready")
	checkpointPath := flag.String("checkpoint", "", "local checkpoint file; required with -apply; rerun the same command to resume")
	flag.Parse()
	project, channel := os.Getenv("GOOGLE_CLOUD_PROJECT"), os.Getenv("LINE_MESSAGING_API_CHANNEL_ID")
	database := os.Getenv("FIRESTORE_DATABASE_ID")
	if database == "" {
		database = "(default)"
	}
	if flag.NArg() != 0 || strings.TrimSpace(project) == "" || !validChannel(channel) {
		return errors.New("set GOOGLE_CLOUD_PROJECT and numeric LINE_MESSAGING_API_CHANNEL_ID; no positional arguments are accepted")
	}
	if *activate && !*apply {
		return errors.New("-activate requires -apply and its completed checkpoint")
	}
	if *apply && *checkpointPath == "" {
		return errors.New("-apply requires a local -checkpoint file for resumable progress")
	}
	saved := checkpoint{Project: project, Database: database}
	if *checkpointPath != "" {
		b, err := os.ReadFile(*checkpointPath)
		if err == nil {
			if json.Unmarshal(b, &saved) != nil {
				return errors.New("invalid migration checkpoint JSON")
			}
			if saved.Project != project || saved.Database != database || saved.Progress.ChannelID != channel || saved.Progress.Apply != *apply {
				return errors.New("checkpoint project/database/channel/mode does not match")
			}
		} else if !errors.Is(err, os.ErrNotExist) {
			return err
		}
	}
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()
	client, err := db.NewClientWithDatabase(ctx, project, database)
	if err != nil {
		return errors.New("cannot initialize Firestore; check explicit project, database and ADC")
	}
	defer func() { resultErr = errors.Join(resultErr, client.Close()) }()
	store := &storage.Attendance{Client: client, ChannelID: channel}
	for !saved.Progress.Complete {
		next, err := store.MigrateQueryBatch(ctx, saved.Progress, *apply)
		if err != nil {
			return err
		}
		saved.Progress = next
		if *checkpointPath != "" {
			b, err := json.MarshalIndent(saved, "", "  ")
			if err != nil {
				return err
			}
			directory := filepath.Dir(*checkpointPath)
			staged, err := os.CreateTemp(directory, ".attendance-migration-*")
			if err != nil {
				return err
			}
			name := staged.Name()
			if _, err = staged.Write(b); err == nil {
				err = staged.Sync()
			}
			closeErr := staged.Close()
			err = errors.Join(err, closeErr)
			if err == nil {
				err = os.Rename(name, *checkpointPath)
			}
			if err != nil {
				_ = os.Remove(name)
				return err
			}
		}
		fmt.Printf("mode=%s records=%d complete=%t\n", map[bool]string{true: "apply", false: "dry-run"}[*apply], next.Records, next.Complete)
	}
	if *activate {
		if err := store.ActivateQueries(ctx, saved.Progress); err != nil {
			return err
		}
		fmt.Println("indexed reads activated")
	}
	return nil
}

func validChannel(channel string) bool {
	if channel == "" {
		return false
	}
	for _, c := range channel {
		if c < '0' || c > '9' {
			return false
		}
	}
	return true
}
