package application

import (
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/963sup/line-mini-app-bot/internal/business/object/application/command"
	"github.com/963sup/line-mini-app-bot/internal/business/object/application/dto"
	"github.com/963sup/line-mini-app-bot/internal/business/object/application/port"
	"github.com/963sup/line-mini-app-bot/internal/business/object/domain/entity"
)

type draftPreparer struct {
	calls      int
	input      DraftPreparation
	candidates []dto.Candidate
	err        error
}

func (p *draftPreparer) PrepareTemplate(_ context.Context, actor string, in DraftPreparation) ([]dto.Candidate, error) {
	p.calls++
	p.input = in
	if actor != "A" {
		panic("untrusted actor")
	}
	return p.candidates, p.err
}

func TestTemplateDraftPreparation(t *testing.T) {
	quantity, unit := "12", "式"
	p := &draftPreparer{candidates: []dto.Candidate{
		{Name: "地板保護", CategoryPath: []string{"電力", "配電"}, Quantity: &quantity, Unit: &unit},
		{Name: "地板保護", CategoryPath: []string{"電力", "接地"}, Quantity: &quantity, Unit: &unit},
		{Name: "搬運（6組）", CategoryPath: []string{"電力", "接地"}, Quantity: &quantity, Unit: &unit},
	}}
	s := &Service{TemplatePreparer: p}
	result, err := s.PrepareTemplateDraft(t.Context(), "A", DraftPreparation{Source: "document", Filename: "工程.txt", Data: []byte("source")})
	if err != nil {
		t.Fatal(err)
	}
	nodes := result.Draft.Nodes
	for index, id := range []string{"n3", "n5", "n6"} {
		got, ok := result.CandidateIndexByNode[id]
		if !ok || got != index {
			t.Fatalf("candidate for %s: %d, %t", id, got, ok)
		}
	}
	if result.Draft.Name != "工程" || len(nodes) != 7 || nodes[3].ParentID == nodes[5].ParentID || nodes[5].ParentID != nodes[6].ParentID {
		t.Fatalf("lost hierarchy: %+v", nodes)
	}
	for _, node := range nodes {
		if node.Quantity != "" {
			t.Fatal("document quantity copied into template")
		}
	}
	if _, err := command.NormalizeTemplate(result.Draft); err != nil {
		t.Fatal(err)
	}
	p.candidates[0].Name = strings.Repeat("長", 121)
	result, err = s.PrepareTemplateDraft(t.Context(), "A", DraftPreparation{Source: "procedure", Name: "防水"})
	if err != nil || len(result.Warnings) != 2 || result.Draft.Nodes[3].Name != p.candidates[0].Name {
		t.Fatalf("must preserve invalid draft for review: %+v %v", result, err)
	}
}

func TestTemplateDraftRejectsBeforeProvider(t *testing.T) {
	for _, in := range []DraftPreparation{{}, {Source: "procedure"}, {Source: "document", Filename: "x"}, {Source: "procedure", Name: "x", Data: []byte("x")}, {Source: "document", Filename: "x", Data: make([]byte, (8<<20)+1)}} {
		p := &draftPreparer{}
		s := &Service{TemplatePreparer: p}
		if _, err := s.PrepareTemplateDraft(t.Context(), "A", in); !errors.Is(err, entity.ErrInput) || p.calls != 0 {
			t.Fatal(err, p.calls)
		}
	}
	p := &draftPreparer{}
	s := &Service{TemplatePreparer: p}
	if _, err := s.PrepareTemplateDraft(t.Context(), "", DraftPreparation{Source: "procedure", Name: "x"}); !errors.Is(err, entity.ErrInput) || p.calls != 0 {
		t.Fatal(err)
	}
	p.err = context.Canceled
	if _, err := s.PrepareTemplateDraft(t.Context(), "A", DraftPreparation{Source: "procedure", Name: "x"}); !errors.Is(err, context.Canceled) {
		t.Fatal(err)
	}
	p.err = nil
	if _, err := s.PrepareTemplateDraft(t.Context(), "A", DraftPreparation{Source: "procedure", Name: "x"}); !errors.Is(err, port.ErrInvalidCandidate) {
		t.Fatal(err)
	}
}
