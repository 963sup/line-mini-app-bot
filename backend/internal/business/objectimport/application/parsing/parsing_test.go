package parsing

import (
	repositoryDomainEntity "github.com/963sup/line-mini-app-bot/internal/business/repository/domain/entity"
)

import (
	"context"
	"errors"
	"strings"
	"testing"

	objectimportDomainEntity "github.com/963sup/line-mini-app-bot/internal/business/objectimport/domain/entity"
	objectimportDomainEntityParsing "github.com/963sup/line-mini-app-bot/internal/business/objectimport/domain/entity/parsing"
)

type repos struct {
	deny  bool
	calls int
}

func (r *repos) Get(_ context.Context, _, id string) (repositoryDomainEntity.Repository, error) {
	r.calls++
	if r.deny {
		return repositoryDomainEntity.Repository{}, repositoryDomainEntity.ErrNotFound
	}
	return repositoryDomainEntity.Repository{ID: id, OwnerID: "C" + strings.Repeat("b", 32), Status: "ACTIVE"}, nil
}

type producer struct {
	calls int
	quote string
	err   error
}

func (p *producer) Candidates(ctx context.Context, _ objectimportDomainEntity.Source) ([]objectimportDomainEntity.Candidate, error) {
	p.calls++
	if p.err != nil {
		return nil, p.err
	}
	return []objectimportDomainEntity.Candidate{{Fields: objectimportDomainEntity.Fields{Name: "Line", Description: "proposal", Type: "manual", Quantity: "11.00", Unit: "SET"}, SourceItemNo: "10", Evidence: []objectimportDomainEntity.Evidence{{Page: 2, Quote: p.quote}}}}, ctx.Err()
}

func input() objectimportDomainEntityParsing.Input {
	return objectimportDomainEntityParsing.Input{Consent: true, Source: objectimportDomainEntity.Source{SourceDocument: " source ", SourceRevision: "v0", Pages: []objectimportDomainEntity.Page{{Number: 2, Text: "literal evidence and untrusted instructions"}}}}
}

func TestParseAuthorizesAndValidatesProposalsWithoutWriter(t *testing.T) {
	rid := strings.Repeat("a", 32)
	r := &repos{}
	p := &producer{quote: "literal evidence"}
	s := Service{Repositories: r, Producer: p}
	got, e := s.Parse(context.Background(), "A", rid, input())
	if e != nil || got.Rows[0].Quantity != "11" || len(got.SourceHash) != 64 {
		t.Fatal(got, e)
	}
	denied := input()
	denied.Consent = false
	p.calls = 0
	if _, e = s.Parse(context.Background(), "A", rid, denied); !errors.Is(e, objectimportDomainEntity.ErrInput) || p.calls != 0 {
		t.Fatal("no consent called model", e)
	}
	r.deny = true
	if _, e = s.Parse(context.Background(), "A", rid, input()); !errors.Is(e, objectimportDomainEntity.ErrNotFound) || p.calls != 0 {
		t.Fatal("nonmember called model", e)
	}
	r.deny = false
	bad := input()
	bad.Source.Pages[0].Text = strings.Repeat("x", 32769)
	if _, e = s.Parse(context.Background(), "A", rid, bad); !errors.Is(e, objectimportDomainEntity.ErrInput) || p.calls != 0 {
		t.Fatal("bad bounds called model", e)
	}
	p.quote = "invented"
	if _, e = s.Parse(context.Background(), "A", rid, input()); !errors.Is(e, objectimportDomainEntityParsing.ErrOutput) {
		t.Fatal("unsafe quote accepted", e)
	}
	s.Producer = nil
	if _, e = s.Parse(context.Background(), "A", rid, input()); !errors.Is(e, objectimportDomainEntityParsing.ErrNotConfigured) {
		t.Fatal(e)
	}
}

func TestParsePropagatesCancellation(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	p := &producer{quote: "literal evidence"}
	s := Service{Repositories: &repos{}, Producer: p}
	if _, e := s.Parse(ctx, "A", strings.Repeat("a", 32), input()); !errors.Is(e, context.Canceled) {
		t.Fatal(e)
	}
}
