package application

import (
	"context"
	"errors"
	"testing"
)

type runStore struct {
	run      Run
	failSave bool
}

func (s *runStore) Load(context.Context, string, string) (Run, error) { return s.run, nil }
func (s *runStore) CompareAndSave(_ context.Context, _, _ string, expected uint64, next Run) error {
	if s.failSave || s.run.Version != expected {
		return ErrConflict
	}
	s.run = next
	return nil
}

type testCommand struct {
	writes map[string]string
	calls  int
	denied bool
}

func (c *testCommand) Execute(_ context.Context, _ string, _ Candidate, key string) (string, error) {
	c.calls++
	if c.denied {
		return "", ErrConflict
	}
	if c.writes[key] == "" {
		c.writes[key] = "journal-receipt"
	}
	return c.writes[key], nil
}

func TestConfirmedWriteCanResumeAfterProgressSaveFails(t *testing.T) {
	ctx := context.Background()
	candidate := Candidate{Version: "1", Target: "repository-b", Command: "CreateJournal"}
	store := &runStore{}
	command := &testCommand{writes: map[string]string{}}
	service := Service{Runs: store, Command: command}
	if _, err := service.Resume(ctx, "actor", "run"); !errors.Is(err, ErrConflict) || command.calls != 0 {
		t.Fatal("unconfirmed write reached command")
	}
	scope := Scope{Actor: "actor", Target: candidate.Target}
	if err := service.Stage(ctx, scope, "run", 0, candidate); err != nil {
		t.Fatal(err)
	}
	stale := candidate
	stale.Version = "outdated"
	if err := service.Stage(ctx, scope, "run", 0, stale); !errors.Is(err, ErrConflict) {
		t.Fatal("old preparation could overwrite the staged candidate")
	}
	changed := candidate
	changed.Target = "repository-c"
	if err := service.Confirm(ctx, "actor", "run", changed, "write-key"); !errors.Is(err, ErrConflict) {
		t.Fatal("confirmation accepted a different target")
	}
	if err := service.Confirm(ctx, "actor", "run", candidate, "write-key"); err != nil {
		t.Fatal(err)
	}
	if _, err := service.Resume(ctx, "other-actor", "run"); !errors.Is(err, ErrConflict) || command.calls != 0 {
		t.Fatal("another actor reused the confirmation")
	}
	command.denied = true
	if _, err := service.Resume(ctx, "actor", "run"); !errors.Is(err, ErrConflict) || len(command.writes) != 0 {
		t.Fatal("workflow bypassed command rejection")
	}
	command.denied = false
	store.failSave = true
	if _, err := service.Resume(ctx, "actor", "run"); !errors.Is(err, ErrConflict) {
		t.Fatal("expected progress write failure")
	}
	store.failSave = false
	if receipt, err := service.Resume(ctx, "actor", "run"); err != nil || receipt != "journal-receipt" || len(command.writes) != 1 {
		t.Fatalf("resume: receipt %q, error %v, unique writes %d", receipt, err, len(command.writes))
	}
	calls := command.calls
	if _, err := service.Resume(ctx, "actor", "run"); err != nil || command.calls != calls {
		t.Fatal("completed workflow called the command again")
	}
}
