package objectimport

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/963sup/line-mini-app-bot/internal/business/repository"
)

func creation() Creation {
	return Creation{SourceDocument: "source", SourceRevision: "v0", Pages: []Page{{Number: 2, Text: "untrusted instructions ignored; item10 quantity11 SET"}}, Rows: []Candidate{{Fields: Fields{Name: "Line item", Description: "Proposal", Type: "manual", Quantity: "11.00", Unit: "SET"}, SourceItemNo: "10", Evidence: []Evidence{{Page: 2, Quote: "item10 quantity11 SET"}}}}}
}

func clone(c Creation) Creation {
	c.Pages = append([]Page{}, c.Pages...)
	c.Rows = append([]Candidate{}, c.Rows...)
	for i := range c.Rows {
		c.Rows[i].Evidence = append([]Evidence{}, c.Rows[i].Evidence...)
	}
	return c
}

func TestPrepareUntrustedCandidates(t *testing.T) {
	good := creation()
	j, hash, e := Prepare(good)
	if e != nil || j.Rows[0].Quantity != "11" || len(hash) != 64 || len(j.SourceHash) != 64 {
		t.Fatalf("normalize: %+v %s %v", j, hash, e)
	}
	// Normalization-equivalent inputs replay the same payload; page text changes do not.
	other := clone(good)
	other.Rows[0].Quantity = "11"
	_, h, e := Prepare(other)
	if e != nil || h != hash {
		t.Fatal("canonical payload mismatch")
	}
	other.Pages[0].Text += " added"
	_, h, e = Prepare(other)
	if e != nil || h == hash {
		t.Fatal("source hash ignored")
	}
	cases := []struct {
		name string
		edit func(*Creation)
	}{
		{"bad quote", func(c *Creation) { c.Rows[0].Evidence[0].Quote = "not in source" }},
		{"wrong page", func(c *Creation) { c.Rows[0].Evidence[0].Page = 3 }},
		{"no evidence", func(c *Creation) { c.Rows[0].Evidence = nil }},
		{"duplicate page", func(c *Creation) { c.Pages = append(c.Pages, c.Pages[0]) }},
		{"duplicate normalized item", func(c *Creation) { c.Rows = append(c.Rows, c.Rows[0]); c.Rows[1].SourceItemNo = " 10 " }},
		{"zero quantity", func(c *Creation) { c.Rows[0].Quantity = "0" }},
		{"too many rows", func(c *Creation) {
			for len(c.Rows) < 101 {
				r := c.Rows[0]
				r.SourceItemNo = strings.Repeat("x", len(c.Rows))
				c.Rows = append(c.Rows, r)
			}
		}},
		{"page byte bound", func(c *Creation) { c.Pages[0].Text = strings.Repeat("字", 11000) }},
		{"invalid utf8", func(c *Creation) { c.Pages[0].Text = "\xff" }},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			c := clone(good)
			tc.edit(&c)
			if _, _, e := Prepare(c); !errors.Is(e, ErrInput) {
				t.Fatal("accepted", e)
			}
		})
	}
}

func draft(t *testing.T) Job {
	t.Helper()
	j, _, e := Prepare(creation())
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
	review := Review{ExpectedVersion: 1, Rows: []ReviewedRow{{ID: j.Rows[0].ID, Fields: j.Rows[0].Fields}}}
	review.Rows[0].Name = "Human correction"
	if _, e := Begin(j, "A", 1, j.UpdatedAt); !errors.Is(e, ErrConflict) {
		t.Fatal("unreviewed confirm", e)
	}
	reviewed, e := ApplyReview(j, "A", review, j.UpdatedAt.Add(time.Second))
	if e != nil || reviewed.Status != "reviewed" || reviewed.Version != 2 || reviewed.Rows[0].SourceItemNo != "10" || reviewed.Rows[0].Evidence[0].Quote != j.Rows[0].Evidence[0].Quote {
		t.Fatal(reviewed, e)
	}
	frozen, e := Begin(reviewed, "A", 2, reviewed.UpdatedAt.Add(time.Second))
	if e != nil || frozen.ConfirmedFromReviewVersion != 2 {
		t.Fatal(frozen, e)
	}
	if _, e = ApplyReview(frozen, "A", Review{ExpectedVersion: frozen.Version, Rows: review.Rows}, frozen.UpdatedAt); !errors.Is(e, ErrConflict) {
		t.Fatal("edited frozen", e)
	}
	if _, e = Begin(frozen, "A", frozen.Version, frozen.UpdatedAt); !errors.Is(e, ErrConflict) {
		t.Fatal("accepted job version instead of review version", e)
	}
	done, e := Record(frozen, "A", j.Rows[0].ID, strings.Repeat("d", 32), 2, frozen.UpdatedAt.Add(time.Second))
	if e != nil || done.Status != "completed" {
		t.Fatal(done, e)
	}
	replay, e := Begin(done, "A", 2, done.UpdatedAt)
	if e != nil || replay.Version != done.Version {
		t.Fatal(replay, e)
	}
	if _, e = Record(done, "A", j.Rows[0].ID, strings.Repeat("e", 32), 2, done.UpdatedAt); !errors.Is(e, ErrConflict) {
		t.Fatal("overwrote result", e)
	}
	if _, e = ApplyReview(j, "B", review, j.UpdatedAt); !errors.Is(e, ErrForbidden) {
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
	input := Review{ExpectedVersion: 1}
	for _, r := range j.Rows {
		f := r.Fields
		f.Description = strings.Repeat("x", 8000)
		input.Rows = append(input.Rows, ReviewedRow{ID: r.ID, Fields: f})
	}
	if _, e := ApplyReview(j, "A", input, j.UpdatedAt); !errors.Is(e, ErrInput) {
		t.Fatal("oversize review", e)
	}
	j = draft(t)
	if e := ValidateStored(j); e != nil {
		t.Fatal(e)
	}
	j.SourceHash = "bad"
	if e := ValidateStored(j); !errors.Is(e, ErrUnavailable) {
		t.Fatal("corrupt source", e)
	}
	j = draft(t)
	j.Rows[0].Evidence[0].Quote = ""
	if e := ValidateStored(j); !errors.Is(e, ErrUnavailable) {
		t.Fatal("corrupt quote", e)
	}
}

type noRepositories struct{}

func (noRepositories) Get(context.Context, string, string) (repository.Repository, error) {
	return repository.Repository{}, repository.ErrNotFound
}

type storeCalls struct{ calls int }

func (s *storeCalls) Create(context.Context, repository.Repository, string, string, string, Job) (Job, error) {
	s.calls++
	return Job{}, nil
}

func (s *storeCalls) Get(context.Context, repository.Repository, string, string) (Job, error) {
	s.calls++
	return Job{}, nil
}

func (s *storeCalls) Review(context.Context, repository.Repository, string, string, Review) (Job, error) {
	s.calls++
	return Job{}, nil
}

func (s *storeCalls) BeginConfirm(context.Context, repository.Repository, string, string, int64) (Job, error) {
	s.calls++
	return Job{}, nil
}

func (s *storeCalls) RecordResult(context.Context, repository.Repository, string, string, string, string, int64) (Job, error) {
	s.calls++
	return Job{}, nil
}

func TestEveryOperationRequiresCurrentRepository(t *testing.T) {
	s := &storeCalls{}
	svc := Service{Repositories: noRepositories{}, Store: s}
	ctx := context.Background()
	rid := strings.Repeat("a", 32)
	_, a := svc.Create(ctx, "A", rid, strings.Repeat("k", 16), creation())
	_, b := svc.Get(ctx, "A", rid, rid)
	_, c := svc.Review(ctx, "A", rid, rid, Review{})
	_, d := svc.Confirm(ctx, "A", rid, rid, 2)
	for _, e := range []error{a, b, c, d} {
		if !errors.Is(e, ErrNotFound) {
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
	n, _ := projectedBytes(j)
	remaining := MaxDraftBytes - 4096 - n - 1024
	for i := range j.Rows {
		n := min(remaining, 8000-len(j.Rows[i].Description))
		j.Rows[i].Description += strings.Repeat("x", n)
		remaining -= n
	}
	if !bounded(j) {
		t.Fatal("near-boundary initial draft rejected")
	}
	j.Rows[0].Description += "x"
	if bounded(j) {
		t.Fatal("boundary padding did not reach last accepted byte")
	}
	j.Rows[0].Description = strings.TrimSuffix(j.Rows[0].Description, "x")
	now := j.UpdatedAt.Add(time.Second)
	review := Review{ExpectedVersion: 1}
	for _, r := range j.Rows {
		review.Rows = append(review.Rows, ReviewedRow{ID: r.ID, Fields: r.Fields})
	}
	j, e := ApplyReview(j, "A", review, now)
	if e != nil {
		t.Fatal(e)
	}
	j, e = Begin(j, "A", 2, now)
	if e != nil {
		t.Fatal(e)
	}
	rows := append([]Row{}, j.Rows...)
	for _, r := range rows {
		j, e = Record(j, "A", r.ID, strings.Repeat("f", 32), 2, now)
		if e != nil {
			t.Fatal(e)
		}
		if e = ValidateStored(j); e != nil {
			t.Fatalf("result did not fit reserved space: %v", e)
		}
	}
	j = draft(t)
	j.Status = "reviewed"
	j.ReviewedAt = &now
	j.Version = repository.MaxResourceNumber - 1
	if _, e = Begin(j, "A", j.Version, now); !errors.Is(e, ErrConflict) {
		t.Fatal("accepted insufficient version budget", e)
	}
}
