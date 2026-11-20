package automatic

import (
	"strings"
	"testing"

	objectimportDomainEntity "github.com/963sup/line-mini-app-bot/internal/business/objectimport/domain/entity"
	objectimportDomainEntityAutomatic "github.com/963sup/line-mini-app-bot/internal/business/objectimport/domain/entity/automatic"
)

func TestAutomaticCandidateRules(t *testing.T) {
	source := objectimportDomainEntity.Source{Pages: []objectimportDomainEntity.Page{{Number: 7, Text: "synthetic"}}}
	j := objectimportDomainEntityAutomatic.Run{ID: strings.Repeat("a", 32), SourceDocument: "synthetic"}
	base := objectimportDomainEntityAutomatic.Candidate{Name: "Pump", Quantity: "11.00", Unit: "SET", SourceItemNo: " 10 ", Pages: []int{7}}
	rows, e := objectimportDomainEntityAutomatic.Prepare(j, source, []objectimportDomainEntityAutomatic.Candidate{base}, "work")
	if e != nil || rows[0].Quantity != "11" || rows[0].SourceItemNo != "10" || rows[0].SourcePages != "7" || rows[0].Type != "work" || rows[0].Description != "" || rows[0].MaterialCode != "" {
		t.Fatal(rows, e)
	}
	cases := []objectimportDomainEntityAutomatic.Candidate{base, base, base, base, base}
	cases[0].Quantity = "0"
	cases[1].Name = ""
	cases[2].Quantity = "NaN"
	cases[3].Unit = ""
	cases[4].Pages = []int{1}
	for _, c := range cases {
		if _, e := objectimportDomainEntityAutomatic.Prepare(j, source, []objectimportDomainEntityAutomatic.Candidate{c}, "work"); e == nil {
			t.Fatal("invalid accepted", c)
		}
	}
	if _, e := objectimportDomainEntityAutomatic.Prepare(j, source, []objectimportDomainEntityAutomatic.Candidate{base, base}, "work"); e == nil {
		t.Fatal("duplicate normalized source item")
	}
}

func TestMixedOrderRequiresItemScopedEvidence(t *testing.T) {
	source := objectimportDomainEntity.Source{Pages: []objectimportDomainEntity.Page{{Number: 1, Text: "[[item:1]] pump supply only\n[[item:2]] pump installation\n[[item:3]] pump"}}}
	candidates := []objectimportDomainEntityAutomatic.Candidate{}
	for _, item := range []struct{ id, kind, quote string }{{"1", "non_work", "supply only"}, {"2", "work", "installation"}, {"3", "work", "installation"}} {
		candidates = append(candidates, objectimportDomainEntityAutomatic.Candidate{Name: "pump", Quantity: "1", Unit: "SET", SourceItemNo: item.id, Type: item.kind, Quote: item.quote, Pages: []int{1}})
	}
	rows, err := objectimportDomainEntityAutomatic.Prepare(objectimportDomainEntityAutomatic.Run{}, source, candidates, "per_item")
	if err != nil || rows[0].Type != "non_work" || rows[1].Type != "work" || rows[2].Type != "needs_review" {
		t.Fatal(rows, err)
	}
	if _, err := objectimportDomainEntityAutomatic.Prepare(objectimportDomainEntityAutomatic.Run{}, source, candidates[:2], "per_item"); err == nil {
		t.Fatal("accepted missing source item")
	}
}

func TestPDFUsesSourceFieldsAndRequiresUnambiguousClassification(t *testing.T) {
	source := objectimportDomainEntity.Source{SourceDocument: "pdf:test", Pages: []objectimportDomainEntity.Page{{
		Number: 1,
		Text:   "[[item:1]] Pump supply only 2 EA\n[[item:2]] Lamp including installation 3 EA\nwithout installation\n[[item:3]] Fan supply only",
	}}, PDF: objectimportDomainEntity.LocatedPDFItems{
		Items: []objectimportDomainEntity.PDFItem{
			{ItemNo: "1", Name: "Pump", Quantity: "2", Unit: "EA"},
			{ItemNo: "2", Name: "Lamp", Quantity: "3", Unit: "EA"},
			{ItemNo: "3", Name: "Fan", Unit: "EA"},
		},
	}}
	candidates := []objectimportDomainEntityAutomatic.Candidate{
		{Name: "Pump", Quantity: "99", Unit: "BOX", SourceItemNo: "1", Type: "non_work", Quote: "supply only", Pages: []int{1}},
		{Name: "Lamp", Quantity: "1", Unit: "BOX", SourceItemNo: "2", Type: "work", Quote: "including installation", Pages: []int{1}},
		{Name: "Fan", Quantity: "10", Unit: "BOX", SourceItemNo: "3", Type: "non_work", Quote: "supply only", Pages: []int{1}},
	}
	rows, err := objectimportDomainEntityAutomatic.Prepare(objectimportDomainEntityAutomatic.Run{SourceDocument: source.SourceDocument}, source, candidates, "per_item")
	if err != nil || len(rows) != 3 {
		t.Fatal(rows, err)
	}
	if rows[0].Quantity != "2" || rows[0].Unit != "EA" || rows[0].Type != "non_work" ||
		rows[1].Type != "needs_review" || rows[2].Quantity != "" || rows[2].Type != "needs_review" {
		t.Fatal(rows)
	}
}

func TestTextOnlyPDFRequiresReviewEvenWithModelClassification(t *testing.T) {
	source := objectimportDomainEntity.Source{SourceDocument: "pdf:test", Pages: []objectimportDomainEntity.Page{{
		Number: 2, Text: "A-10 Pump supply only\n2 EA",
	}}, PDF: objectimportDomainEntity.PDFPageText{}}
	candidate := objectimportDomainEntityAutomatic.Candidate{
		Name: "Pump", Quantity: "2", Unit: "EA", SourceItemNo: "A-10",
		Type: "non_work", Quote: "Pump supply only 2 EA", Pages: []int{2},
	}
	run := objectimportDomainEntityAutomatic.Run{SourceDocument: source.SourceDocument}
	rows, err := objectimportDomainEntityAutomatic.Prepare(run, source, []objectimportDomainEntityAutomatic.Candidate{candidate}, "per_item")
	if err != nil || len(rows) != 1 || rows[0].Type != "needs_review" {
		t.Fatal(rows, err)
	}
	candidate.Quote = "not in source"
	if _, err := objectimportDomainEntityAutomatic.Prepare(run, source, []objectimportDomainEntityAutomatic.Candidate{candidate}, "per_item"); err == nil {
		t.Fatal("accepted an unverified quotation")
	}
}
