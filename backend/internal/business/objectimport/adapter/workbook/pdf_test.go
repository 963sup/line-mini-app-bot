package workbook

import (
	"os"
	"testing"
)

func TestOrderAfterCover(t *testing.T) {
	data, err := os.ReadFile("testdata/cover-order.pdf")
	if err != nil {
		t.Fatal(err)
	}
	pages, err := ReadPDF(t.Context(), data)
	if err != nil || len(pages) != 1 || pages[0].Number != 2 || pages[0].Text != "[[item:001]] 001 工程 20 組" {
		t.Fatalf("pages=%+v err=%v", pages, err)
	}
	items := pages[0].PDFItems
	if len(items) != 1 || items[0].ItemNo != "001" || items[0].Quantity != "20" || items[0].Unit != "組" {
		t.Fatalf("unexpected source fields: %+v", items)
	}
}

func TestTextPDFWithoutRecognizedTable(t *testing.T) {
	data, err := os.ReadFile("testdata/plain-order.pdf")
	if err != nil {
		t.Fatal(err)
	}
	pages, err := ReadPDF(t.Context(), data)
	if err != nil || len(pages) != 1 || !pages[0].PDFTextOnly ||
		pages[0].Text != "Order details\nA-10 Pump supply only 2 EA\nB-20 Lamp including installation" {
		t.Fatalf("pages=%+v err=%v", pages, err)
	}
}
