package application

import (
	repositoryDomainEntity "github.com/963sup/line-mini-app-bot/internal/business/repository/domain/entity"
)

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	objectimportDomainEntity "github.com/963sup/line-mini-app-bot/internal/business/objectimport/domain/entity"
)

func creation() objectimportDomainEntity.Creation {
	return objectimportDomainEntity.Creation{SourceDocument: "source", SourceRevision: "v0", Pages: []objectimportDomainEntity.Page{{Number: 2, Text: "untrusted instructions ignored; item10 quantity11 SET"}}, Rows: []objectimportDomainEntity.Candidate{{Fields: objectimportDomainEntity.Fields{Name: "Line item", Description: "Proposal", Type: "manual", Quantity: "11.00", Unit: "SET"}, SourceItemNo: "10", Evidence: []objectimportDomainEntity.Evidence{{Page: 2, Quote: "item10 quantity11 SET"}}}}}
}

func clone(c objectimportDomainEntity.Creation) objectimportDomainEntity.Creation {
	c.Pages = append([]objectimportDomainEntity.Page{}, c.Pages...)
	c.Rows = append([]objectimportDomainEntity.Candidate{}, c.Rows...)
	for i := range c.Rows {
		c.Rows[i].Evidence = append([]objectimportDomainEntity.Evidence{}, c.Rows[i].Evidence...)
	}
	return c
}

func TestPrepareUntrustedCandidates(t *testing.T) {
	good := creation()
	j, hash, e := objectimportDomainEntity.Prepare(good)
	if e != nil || j.Rows[0].Quantity != "11" || len(hash) != 64 || len(j.SourceHash) != 64 {
		t.Fatalf("normalize: %+v %s %v", j, hash, e)
	}
	// Normalization-equivalent inputs replay the same payload; page text changes do not.
	other := clone(good)
	other.Rows[0].Quantity = "11"
	_, h, e := objectimportDomainEntity.Prepare(other)
	if e != nil || h != hash {
		t.Fatal("canonical payload mismatch")
	}
	other.Pages[0].Text += " added"
	_, h, e = objectimportDomainEntity.Prepare(other)
	if e != nil || h == hash {
		t.Fatal("source hash ignored")
	}
	cases := []struct {
		name string
		edit func(*objectimportDomainEntity.Creation)
	}{
		{"bad quote", func(c *objectimportDomainEntity.Creation) { c.Rows[0].Evidence[0].Quote = "not in source" }},
		{"wrong page", func(c *objectimportDomainEntity.Creation) { c.Rows[0].Evidence[0].Page = 3 }},
		{"no evidence", func(c *objectimportDomainEntity.Creation) { c.Rows[0].Evidence = nil }},
		{"duplicate page", func(c *objectimportDomainEntity.Creation) { c.Pages = append(c.Pages, c.Pages[0]) }},
		{"duplicate normalized item", func(c *objectimportDomainEntity.Creation) {
			c.Rows = append(c.Rows, c.Rows[0])
			c.Rows[1].SourceItemNo = " 10 "
		}},
		{"zero quantity", func(c *objectimportDomainEntity.Creation) { c.Rows[0].Quantity = "0" }},
		{"too many rows", func(c *objectimportDomainEntity.Creation) {
			for len(c.Rows) < 101 {
				r := c.Rows[0]
				r.SourceItemNo = strings.Repeat("x", len(c.Rows))
				c.Rows = append(c.Rows, r)
			}
		}},
		{"page byte bound", func(c *objectimportDomainEntity.Creation) { c.Pages[0].Text = strings.Repeat("字", 11000) }},
		{"invalid utf8", func(c *objectimportDomainEntity.Creation) { c.Pages[0].Text = "\xff" }},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			c := clone(good)
			tc.edit(&c)
			if _, _, e := objectimportDomainEntity.Prepare(c); !errors.Is(e, objectimportDomainEntity.ErrInput) {
				t.Fatal("accepted", e)
			}
		})
	}
}

func draft(t *testing.T) objectimportDomainEntity.Job {
	t.Helper()
	j, _, e := objectimportDomainEntity.Prepare(creation())
	if e != nil {
		t.Fatal(e)
	}
	j.ID = strings.Repeat("a", 32)
	j.RepositoryID = strings.Repeat("b", 32)
	j.CreatorID = "A"
	j.CreatedAt = time.Now().UTC()
	j.UpdatedAt = j.CreatedAt
	j.Rows[0].ID = strings.Repeat("c", 32)
	return j
}

func TestReviewFreezeAndResultState(t *testing.T) {
	j := draft(t)
	review := objectimportDomainEntity.Review{ExpectedVersion: 1, Rows: []objectimportDomainEntity.ReviewedRow{{ID: j.Rows[0].ID, Fields: j.Rows[0].Fields}}}
	review.Rows[0].Name = "Human correction"
	if _, e := objectimportDomainEntity.Begin(j, "A", 1, j.UpdatedAt); !errors.Is(e, objectimportDomainEntity.ErrConflict) {
		t.Fatal("unreviewed confirm", e)
	}
	reviewed, e := objectimportDomainEntity.ApplyReview(j, "A", review, j.UpdatedAt.Add(time.Second))
	if e != nil || reviewed.Status != "reviewed" || reviewed.Version != 2 || reviewed.Rows[0].SourceItemNo != "10" || reviewed.Rows[0].Evidence[0].Quote != j.Rows[0].Evidence[0].Quote {
		t.Fatal(reviewed, e)
	}
	frozen, e := objectimportDomainEntity.Begin(reviewed, "A", 2, reviewed.UpdatedAt.Add(time.Second))
	if e != nil || frozen.ConfirmedFromReviewVersion != 2 {
		t.Fatal(frozen, e)
	}
	if _, e = objectimportDomainEntity.ApplyReview(frozen, "A", objectimportDomainEntity.Review{ExpectedVersion: frozen.Version, Rows: review.Rows}, frozen.UpdatedAt); !errors.Is(e, objectimportDomainEntity.ErrConflict) {
		t.Fatal("edited frozen", e)
	}
	if _, e = objectimportDomainEntity.Begin(frozen, "A", frozen.Version, frozen.UpdatedAt); !errors.Is(e, objectimportDomainEntity.ErrConflict) {
		t.Fatal("accepted job version instead of review version", e)
	}
	done, e := objectimportDomainEntity.Record(frozen, "A", j.Rows[0].ID, strings.Repeat("d", 32), 2, frozen.UpdatedAt.Add(time.Second))
	if e != nil || done.Status != "completed" {
		t.Fatal(done, e)
	}
	replay, e := objectimportDomainEntity.Begin(done, "A", 2, done.UpdatedAt)
	if e != nil || replay.Version != done.Version {
		t.Fatal(replay, e)
	}
	if _, e = objectimportDomainEntity.Record(done, "A", j.Rows[0].ID, strings.Repeat("e", 32), 2, done.UpdatedAt); !errors.Is(e, objectimportDomainEntity.ErrConflict) {
		t.Fatal("overwrote result", e)
	}
	if _, e = objectimportDomainEntity.ApplyReview(j, "B", review, j.UpdatedAt); !errors.Is(e, objectimportDomainEntity.ErrForbidden) {
		t.Fatal("owner", e)
	}
}

func TestReviewAggregateAndStoredCorruption(t *testing.T) {
	j := draft(t)
	for i := 1; i < 40; i++ {
		r := j.Rows[0]
		r.ID = strings.Repeat("d", 30) + string("0123456789abcdef"[i/16]) + string("0123456789abcdef"[i%16])
		r.SourceItemNo = r.ID
		j.Rows = append(j.Rows, r)
	}
	input := objectimportDomainEntity.Review{ExpectedVersion: 1}
	for _, r := range j.Rows {
		f := r.Fields
		f.Description = strings.Repeat("x", 8000)
		input.Rows = append(input.Rows, objectimportDomainEntity.ReviewedRow{ID: r.ID, Fields: f})
	}
	if _, e := objectimportDomainEntity.ApplyReview(j, "A", input, j.UpdatedAt); !errors.Is(e, objectimportDomainEntity.ErrInput) {
		t.Fatal("oversize review", e)
	}
	j = draft(t)
	if e := objectimportDomainEntity.ValidateStored(j); e != nil {
		t.Fatal(e)
	}
	j.SourceHash = "bad"
	if e := objectimportDomainEntity.ValidateStored(j); !errors.Is(e, objectimportDomainEntity.ErrUnavailable) {
		t.Fatal("corrupt source", e)
	}
	j = draft(t)
	j.Rows[0].Evidence[0].Quote = ""
	if e := objectimportDomainEntity.ValidateStored(j); !errors.Is(e, objectimportDomainEntity.ErrUnavailable) {
		t.Fatal("corrupt quote", e)
	}
}

type noRepositories struct{}

func (noRepositories) Get(context.Context, string, string) (repositoryDomainEntity.Repository, error) {
	return repositoryDomainEntity.Repository{}, repositoryDomainEntity.ErrNotFound
}

type storeCalls struct{ calls int }

func (s *storeCalls) Create(context.Context, repositoryDomainEntity.Repository, string, string, string, objectimportDomainEntity.Job) (objectimportDomainEntity.Job, error) {
	s.calls++
	return objectimportDomainEntity.Job{}, nil
}

func (s *storeCalls) Get(context.Context, repositoryDomainEntity.Repository, string, string) (objectimportDomainEntity.Job, error) {
	s.calls++
	return objectimportDomainEntity.Job{}, nil
}

func (s *storeCalls) Review(context.Context, repositoryDomainEntity.Repository, string, string, objectimportDomainEntity.Review) (objectimportDomainEntity.Job, error) {
	s.calls++
	return objectimportDomainEntity.Job{}, nil
}

func (s *storeCalls) BeginConfirm(context.Context, repositoryDomainEntity.Repository, string, string, int64) (objectimportDomainEntity.Job, error) {
	s.calls++
	return objectimportDomainEntity.Job{}, nil
}

func (s *storeCalls) RecordResult(context.Context, repositoryDomainEntity.Repository, string, string, string, string, int64) (objectimportDomainEntity.Job, error) {
	s.calls++
	return objectimportDomainEntity.Job{}, nil
}

func TestEveryOperationRequiresCurrentRepository(t *testing.T) {
	s := &storeCalls{}
	svc := Service{Repositories: noRepositories{}, Store: s}
	ctx := context.Background()
	rid := strings.Repeat("a", 32)
	_, a := svc.Create(ctx, "A", rid, strings.Repeat("k", 16), creation())
	_, b := svc.Get(ctx, "A", rid, rid)
	_, c := svc.Review(ctx, "A", rid, rid, objectimportDomainEntity.Review{})
	_, d := svc.Confirm(ctx, "A", rid, rid, 2)
	for _, e := range []error{a, b, c, d} {
		if !errors.Is(e, objectimportDomainEntity.ErrNotFound) {
			t.Fatal(e)
		}
	}
	if s.calls != 0 {
		t.Fatal("denied call touched draft store")
	}
}

func TestConfirmationReservesResultSpaceAndVersionBudget(t *testing.T) {
	j := draft(t)
	for i := 1; i < 40; i++ {
		r := j.Rows[0]
		r.ID = strings.Repeat("e", 30) + string("0123456789abcdef"[i/16]) + string("0123456789abcdef"[i%16])
		r.SourceItemNo = r.ID
		j.Rows = append(j.Rows, r)
	}
	n, _ := objectimportDomainEntity.ProjectedBytes(j)
	remaining := objectimportDomainEntity.MaxDraftBytes - 4096 - n - 1024
	for i := range j.Rows {
		n := min(remaining, 8000-len(j.Rows[i].Description))
		j.Rows[i].Description += strings.Repeat("x", n)
		remaining -= n
	}
	if !objectimportDomainEntity.WithinSizeLimit(j) {
		t.Fatal("near-boundary initial draft rejected")
	}
	j.Rows[0].Description += "x"
	if objectimportDomainEntity.WithinSizeLimit(j) {
		t.Fatal("boundary padding did not reach last accepted byte")
	}
	j.Rows[0].Description = strings.TrimSuffix(j.Rows[0].Description, "x")
	now := j.UpdatedAt.Add(time.Second)
	review := objectimportDomainEntity.Review{ExpectedVersion: 1}
	for _, r := range j.Rows {
		review.Rows = append(review.Rows, objectimportDomainEntity.ReviewedRow{ID: r.ID, Fields: r.Fields})
	}
	j, e := objectimportDomainEntity.ApplyReview(j, "A", review, now)
	if e != nil {
		t.Fatal(e)
	}
	j, e = objectimportDomainEntity.Begin(j, "A", 2, now)
	if e != nil {
		t.Fatal(e)
	}
	rows := append([]objectimportDomainEntity.Row{}, j.Rows...)
	for _, r := range rows {
		j, e = objectimportDomainEntity.Record(j, "A", r.ID, strings.Repeat("f", 32), 2, now)
		if e != nil {
			t.Fatal(e)
		}
		if e = objectimportDomainEntity.ValidateStored(j); e != nil {
			t.Fatalf("result did not fit reserved space: %v", e)
		}
	}
	j = draft(t)
	j.Status = "reviewed"
	j.ReviewedAt = &now
	j.Version = repositoryDomainEntity.MaxResourceNumber - 1
	if _, e = objectimportDomainEntity.Begin(j, "A", j.Version, now); !errors.Is(e, objectimportDomainEntity.ErrConflict) {
		t.Fatal("accepted insufficient version budget", e)
	}
}
