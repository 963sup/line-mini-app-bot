package workbook

import (
	"os"
	"testing"

	imp "github.com/963sup/line-mini-app-bot/internal/business/objectimport/domain/entity"
)

func TestOrderAfterCover(t *testing.T) {
	data, err := os.ReadFile("testdata/cover-order.pdf")
	if err != nil {
		t.Fatal(err)
	}
	source, err := ReadPDF(t.Context(), data)
	pages := source.Pages
	if err != nil || len(pages) != 1 || pages[0].Number != 2 || pages[0].Text != "[[item:001]] 001 工程 20 組" {
		t.Fatalf("pages=%+v err=%v", pages, err)
	}
	located, ok := source.Evidence.(imp.LocatedPDFItems)
	if !ok {
		t.Fatalf("unexpected PDF evidence: %T", source.Evidence)
	}
	items := located.Items
	if len(items) != 1 || items[0].ItemNo != "001" || items[0].Quantity != "20" || items[0].Unit != "組" {
		t.Fatalf("unexpected source fields: %+v", items)
	}
}

func TestTextPDFWithoutRecognizedTable(t *testing.T) {
	data, err := os.ReadFile("testdata/plain-order.pdf")
	if err != nil {
		t.Fatal(err)
	}
	source, err := ReadPDF(t.Context(), data)
	pages := source.Pages
	_, textOnly := source.Evidence.(imp.PDFPageText)
	if err != nil || len(pages) != 1 || !textOnly ||
		pages[0].Text != "Order details\nA-10 Pump supply only 2 EA\nB-20 Lamp including installation" {
		t.Fatalf("pages=%+v err=%v", pages, err)
	}
}
