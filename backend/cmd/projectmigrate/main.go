// projectmigrate copies one legacy Project scope with an explicit writer lock.
// Dry-run is the default. It never deploys or selects production credentials.
package main

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"os/signal"
	"path/filepath"
	"regexp"
	"strings"
	"syscall"
	"time"

	db "cloud.google.com/go/firestore"
)

var (
	idPattern      = regexp.MustCompile(`^[a-f0-9]{32}$`)
	groupPattern   = regexp.MustCompile(`^C[a-f0-9]{32}$`)
	channelPattern = regexp.MustCompile(`^[0-9]+$`)
)

type options struct {
	Project, Database, Channel, Repository, Manifest string
	Apply, Rollback, ReleaseLock, WritersReady       bool
	Timeout                                          time.Duration
}

type record struct {
	Source, Target, SourceHash, TargetHash string
}

type manifest struct {
	SchemaVersion                                                       int
	MigrationID, Project, Database, Channel, Repository, SpaceID, Phase string
	Records                                                             []record
}

func main() {
	if err := run(os.Args[1:], os.Stdout); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func parse(args []string) (options, error) {
	var o options
	f := flag.NewFlagSet("projectmigrate", flag.ContinueOnError)
	f.StringVar(&o.Project, "project", "", "explicit Google Cloud project (or emulator project)")
	f.StringVar(&o.Database, "database", "", "explicit Firestore database; use (default) explicitly")
	f.StringVar(&o.Channel, "channel", "", "numeric Messaging API channel scope")
	f.StringVar(&o.Repository, "repository", "", "one legacy source repository ID")
	f.StringVar(&o.Manifest, "manifest", "", "local reviewable manifest; required in every mode")
	f.BoolVar(&o.Apply, "apply", false, "lock writers, copy and verify, then switch the single write location")
	f.BoolVar(&o.Rollback, "rollback", false, "switch back only if no source or target drift; never delete data")
	f.BoolVar(&o.ReleaseLock, "release-lock", false, "release this manifest lock without switching location or changing data")
	f.BoolVar(&o.WritersReady, "writers-ready", false, "confirm every deployed Project writer honors projectStorage locks")
	f.DurationVar(&o.Timeout, "timeout", 15*time.Minute, "overall operation deadline")
	if err := f.Parse(args); err != nil {
		return o, err
	}
	if f.NArg() != 0 || strings.TrimSpace(o.Project) == "" || strings.TrimSpace(o.Database) == "" || !channelPattern.MatchString(o.Channel) || !idPattern.MatchString(o.Repository) || o.Manifest == "" || o.Timeout <= 0 {
		return o, errors.New("explicit -project, -database, -channel, -repository and -manifest are required; no positional arguments")
	}
	modes := 0
	for _, enabled := range []bool{o.Apply, o.Rollback, o.ReleaseLock} {
		if enabled {
			modes++
		}
	}
	if modes > 1 {
		return o, errors.New("choose only one mutating mode")
	}
	if modes > 0 && !o.WritersReady {
		return o, errors.New("mutating mode requires -writers-ready; upgrade or stop all legacy writers first")
	}
	return o, nil
}

func run(args []string, output io.Writer) (resultErr error) {
	o, err := parse(args)
	if err != nil {
		return err
	}
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()
	ctx, cancel := context.WithTimeout(ctx, o.Timeout)
	defer cancel()
	client, err := db.NewClientWithDatabase(ctx, o.Project, o.Database)
	if err != nil {
		return errors.New("cannot initialize explicit Firestore project/database; check ADC or emulator")
	}
	defer func() { resultErr = errors.Join(resultErr, client.Close()) }()
	m := &migrator{client: client, options: o}
	if !o.Apply && !o.Rollback && !o.ReleaseLock {
		if _, err := os.Stat(o.Manifest); err == nil {
			return errors.New("manifest already exists; use a new path to preserve previous evidence")
		} else if !errors.Is(err, os.ErrNotExist) {
			return err
		}
		plan, err := m.plan(ctx)
		if err != nil {
			return err
		}
		if err := save(o.Manifest, plan); err != nil {
			return err
		}
		_, err = fmt.Fprintf(output, "mode=dry-run documents=%d space=%s manifest=%s no_database_writes=true\n", len(plan.Records), plan.SpaceID, o.Manifest)
		return err
	}
	plan, err := load(o.Manifest, o)
	if err != nil {
		return err
	}
	if o.ReleaseLock {
		err = m.releaseLock(ctx, plan)
	} else if o.Rollback {
		err = m.rollback(ctx, &plan)
	} else {
		err = m.apply(ctx, &plan)
	}
	if err != nil {
		return err
	}
	_, err = fmt.Fprintf(output, "phase=%s documents=%d manifest=%s\n", plan.Phase, len(plan.Records), o.Manifest)
	return err
}

func save(path string, value manifest) error {
	b, err := json.MarshalIndent(value, "", "  ")
	if err != nil {
		return err
	}
	f, err := os.CreateTemp(filepath.Dir(path), ".project-migration-*")
	if err != nil {
		return err
	}
	name := f.Name()
	defer func() { _ = os.Remove(name) }()
	if err = f.Chmod(0o600); err == nil {
		_, err = f.Write(append(b, '\n'))
	}
	if err == nil {
		err = f.Sync()
	}
	err = errors.Join(err, f.Close())
	if err != nil {
		return err
	}
	return os.Rename(name, path)
}

func load(path string, o options) (manifest, error) {
	var m manifest
	b, err := os.ReadFile(path)
	if err != nil {
		return m, err
	}
	if json.Unmarshal(b, &m) != nil || m.SchemaVersion != 1 || m.Project != o.Project || m.Database != o.Database || m.Channel != o.Channel || m.Repository != o.Repository || !groupPattern.MatchString(m.SpaceID) || !idPattern.MatchString(m.MigrationID) {
		return m, errors.New("manifest schema or explicit namespace does not match")
	}
	if m.Phase != "planned" && m.Phase != "locked" && m.Phase != "active" && m.Phase != "rolled_back" {
		return m, errors.New("invalid manifest phase")
	}
	seen := map[string]bool{}
	for _, r := range m.Records {
		target, _, err := targetPath(o.Channel, o.Repository, m.SpaceID, r.Source)
		if err != nil || r.Target != target || seen[r.Target] || len(r.SourceHash) != 64 || len(r.TargetHash) != 64 {
			return m, errors.New("manifest contains invalid, duplicate or out-of-scope records")
		}
		seen[r.Target] = true
	}
	return m, nil
}

func digest(data map[string]any) (string, error) {
	b, err := json.Marshal(data)
	if err != nil {
		return "", errors.New("document cannot be represented in migration hash")
	}
	sum := sha256.Sum256(b)
	return hex.EncodeToString(sum[:]), nil
}

func newID() (string, error) {
	var b [16]byte
	if _, err := rand.Read(b[:]); err != nil {
		return "", err
	}
	return hex.EncodeToString(b[:]), nil
}
