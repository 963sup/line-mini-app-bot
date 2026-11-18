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
