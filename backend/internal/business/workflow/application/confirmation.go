// Package application coordinates confirmed writes without owning business authorization.
package application

import (
	"context"
	"errors"
)

var (
	ErrConflict = errors.New("workflow confirmation changed or is unavailable")
	ErrReceipt  = errors.New("workflow command returned no receipt")
)

// Candidate identifies an immutable snapshot and the action it permits.
type Candidate struct {
	Version, Target, Command string
}

// Scope comes from a trusted caller that has already established identity.
type Scope struct {
	Actor, Target string
}

type Run struct {
	Version     uint64
	Candidate   Candidate
	WriteKey    string
	ConfirmedBy string
	Receipt     string
}

// RunStore scopes access to the actor. CompareAndSave must atomically reject
// stale versions and preserve confirmed candidates until the command resolves.
type RunStore interface {
	Load(ctx context.Context, actor, runID string) (Run, error)
	CompareAndSave(ctx context.Context, actor, runID string, expected uint64, next Run) error
}

// Command executes the exact confirmed snapshot, rechecking current business
// authorization. Reusing writeKey must return the original receipt on replay.
type Command interface {
	Execute(ctx context.Context, actor string, candidate Candidate, writeKey string) (receipt string, err error)
}

type Service struct {
	Runs    RunStore
	Command Command
}

// Stage replaces a candidate only when the caller still holds the current version.
func (s Service) Stage(ctx context.Context, scope Scope, runID string, version uint64, candidate Candidate) error {
	if scope.Actor == "" || runID == "" || scope.Target == "" || candidate.Target != scope.Target ||
		candidate.Version == "" || candidate.Command == "" {
		return ErrConflict
	}
	run, err := s.Runs.Load(ctx, scope.Actor, runID)
	if err != nil {
		return err
	}
	if run.Version != version || run.WriteKey != "" || run.Receipt != "" {
		return ErrConflict
	}
	run.Version++
	run.Candidate = candidate
	return s.Runs.CompareAndSave(ctx, scope.Actor, runID, version, run)
}

func (s Service) Confirm(ctx context.Context, actor, runID string, candidate Candidate, writeKey string) error {
	if actor == "" || runID == "" || candidate.Version == "" || candidate.Target == "" || candidate.Command == "" || writeKey == "" {
		return ErrConflict
	}
	run, err := s.Runs.Load(ctx, actor, runID)
	if err != nil {
		return err
	}
	if run.Candidate != candidate || run.WriteKey != "" || run.Receipt != "" {
		return ErrConflict
	}
	expected := run.Version
	run.Version++
	run.WriteKey = writeKey
	run.ConfirmedBy = actor
	return s.Runs.CompareAndSave(ctx, actor, runID, expected, run)
}

func (s Service) Resume(ctx context.Context, actor, runID string) (string, error) {
	if actor == "" || runID == "" {
		return "", ErrConflict
	}
	run, err := s.Runs.Load(ctx, actor, runID)
	if err != nil {
		return "", err
	}
	if run.ConfirmedBy != actor || run.WriteKey == "" {
		return "", ErrConflict
	}
	if run.Receipt != "" {
		return run.Receipt, nil
	}
	receipt, err := s.Command.Execute(ctx, actor, run.Candidate, run.WriteKey)
	if err != nil {
		return "", err
	}
	if receipt == "" {
		return "", ErrReceipt
	}
	expected := run.Version
	run.Version++
	run.Receipt = receipt
	if err := s.Runs.CompareAndSave(ctx, actor, runID, expected, run); err != nil {
		return "", err
	}
	return receipt, nil
}
