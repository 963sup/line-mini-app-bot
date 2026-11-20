package workbook

import "testing"

func TestStructuredTextUsesContractColumnAndPhysicalLine(t *testing.T) {
	for _, tc := range []struct{ format, text, line string }{
		{"csv", "\uFEFF項次,名稱,累計完成數量,契約數量,單位\r\n\r\n01,\"Pump, A\",30,20,SET", "L3"},
		{"txt", "項次\t品名\t數量\t單位\n01\tPump, A\t20\tSET", "L2"},
		{"tsv", "項次\t品名\t數量\t單位\n01\tPump, A\t20\tSET", "L2"},
		{"md", "|項次|名稱|契約數量|累計完成數量|單位|\n|---|---|---:|---:|---|\n|01|Pump, A|20|30|SET|", "L3"},
	} {
		rows, err := ReadText(tc.format, []byte(tc.text))
		if err != nil || len(rows) != 1 {
			t.Fatal(tc.format, err)
		}
		r := rows[0]
		if r.Quantity != "20" || r.Name != "Pump, A" || r.SourceItemNo != "01" || r.Type != "needs_review" || r.SourcePages != tc.line {
			t.Fatal(r)
		}
	}
	for _, text := range []string{"自由文字", "項次,名稱,數量,契約數量,單位\n1,Pump,20,30,SET", "項次,名稱,數量,單位\n1,Pump,20,SET\n1,X,1,SET"} {
		if _, err := ReadText("csv", []byte(text)); err == nil {
			t.Fatal("accepted invalid table")
		}
	}
}
