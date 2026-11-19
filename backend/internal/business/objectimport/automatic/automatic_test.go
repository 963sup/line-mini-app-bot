package automatic

import (
	"strings"
	"testing"

	feature "github.com/963sup/line-mini-app-bot/internal/business/objectimport"
)

func TestAutomaticCandidateRules(t *testing.T) {
	source := feature.Source{Pages: []feature.Page{{Number: 7, Text: "synthetic"}}}
	j := Run{ID: strings.Repeat("a", 32), SourceDocument: "synthetic"}
	base := Candidate{Name: "Pump", Quantity: "11.00", Unit: "SET", SourceItemNo: " 10 ", Pages: []int{7}}
	rows, e := Prepare(j, source, []Candidate{base}, "work")
	if e != nil || rows[0].Quantity != "11" || rows[0].SourceItemNo != "10" || rows[0].SourcePages != "7" || rows[0].Type != "work" || rows[0].Description != "" || rows[0].MaterialCode != "" {
		t.Fatal(rows, e)
	}
	cases := []Candidate{base, base, base, base, base}
	cases[0].Quantity = "0"
	cases[1].Name = ""
	cases[2].Quantity = "NaN"
	cases[3].Unit = ""
	cases[4].Pages = []int{1}
	for _, c := range cases {
		if _, e := Prepare(j, source, []Candidate{c}, "work"); e == nil {
			t.Fatal("invalid accepted", c)
		}
	}
	if _, e := Prepare(j, source, []Candidate{base, base}, "work"); e == nil {
		t.Fatal("duplicate normalized source item")
	}
}

func TestMixedOrderRequiresItemScopedEvidence(t *testing.T) {
	source := feature.Source{Pages: []feature.Page{{Number: 1, Text: "[[item:1]] pump supply only\n[[item:2]] pump installation\n[[item:3]] pump"}}}
	candidates := []Candidate{}
	for _, item := range []struct{ id, kind, quote string }{{"1", "non_work", "supply only"}, {"2", "work", "installation"}, {"3", "work", "installation"}} {
		candidates = append(candidates, Candidate{Name: "pump", Quantity: "1", Unit: "SET", SourceItemNo: item.id, Type: item.kind, Quote: item.quote, Pages: []int{1}})
	}
	rows, err := Prepare(Run{}, source, candidates, "per_item")
	if err != nil || rows[0].Type != "non_work" || rows[1].Type != "work" || rows[2].Type != "needs_review" {
		t.Fatal(rows, err)
	}
	if _, err := Prepare(Run{}, source, candidates[:2], "per_item"); err == nil {
		t.Fatal("accepted missing source item")
	}
}
