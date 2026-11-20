package workbook

import (
	"bytes"
	"strings"

	object "github.com/963sup/line-mini-app-bot/internal/business/object/domain/entity"
	imp "github.com/963sup/line-mini-app-bot/internal/business/objectimport/domain/entity"
	run "github.com/963sup/line-mini-app-bot/internal/business/objectimport/domain/entity/automatic"
	"github.com/xuri/excelize/v2"
)

// Read accepts literal detail rows from one explicitly headed worksheet.
func Read(data []byte) ([]run.Row, error) {
	f, err := excelize.OpenReader(bytes.NewReader(data), excelize.Options{UnzipSizeLimit: 16 << 20, UnzipXMLSizeLimit: 4 << 20, RawCellValue: true})
	if err != nil {
		return nil, imp.ErrInput
	}
	defer func() { _ = f.Close() }()
	sheets := f.GetSheetList()
	if len(sheets) != 1 {
		return nil, imp.ErrInput
	}
	sheet := sheets[0]
	it, err := f.Rows(sheet)
	if err != nil {
		return nil, imp.ErrInput
	}
	defer func() { _ = it.Close() }()
	if !it.Next() {
		return nil, imp.ErrInput
	}
	header, err := it.Columns()
	if err != nil {
		return nil, imp.ErrInput
	}
	cols, err := columns(header)
	if err != nil {
		return nil, err
	}
	var out []run.Row
	seen := map[string]bool{}
	for number := 2; it.Next(); number++ {
		if number > 10000 {
			return nil, imp.ErrInput
		}
		values, err := it.Columns()
		if err != nil {
			return nil, imp.ErrInput
		}
		if len(values) == 0 {
			continue
		}
		selected := map[string]string{}
		var location string
		for field, col := range cols {
			if col >= len(values) {
				return nil, imp.ErrInput
			}
			cell, _ := excelize.CoordinatesToCellName(col+1, number)
			formula, err := f.GetCellFormula(sheet, cell)
			if err != nil || formula != "" {
				return nil, imp.ErrInput
			}
			selected[field] = strings.TrimSpace(values[col])
			if field == "quantity" {
				location = "'" + strings.ReplaceAll(sheet, "'", "''") + "'!" + cell
			}
		}
		if err := appendRow(&out, seen, selected, location); err != nil {
			return nil, err
		}
	}
	if len(out) == 0 || it.Error() != nil {
		return nil, imp.ErrInput
	}
	return out, nil
}

func columns(header []string) (map[string]int, error) {
	cols := map[string]int{}
	for i, h := range header {
		switch strings.TrimSpace(h) {
		case "項次":
			h = "item"
		case "品名", "料號/品名", "名稱":
			h = "name"
		case "數量", "契約數量":
			h = "quantity"
		case "單位":
			h = "unit"
		default:
			continue
		}
		if _, dup := cols[h]; dup {
			return nil, imp.ErrInput
		}
		cols[h] = i
	}
	if len(cols) != 4 {
		return nil, imp.ErrInput
	}
	return cols, nil
}

func appendRow(out *[]run.Row, seen map[string]bool, selected map[string]string, location string) error {
	content, err := object.Normalize(object.Content{Name: selected["name"], Type: "needs_review", Quantity: selected["quantity"], Unit: selected["unit"], SourceItemNo: selected["item"], SourcePages: location})
	if err != nil || content.Name == "" || content.SourceItemNo == "" || content.Unit == "" || selected["quantity"] == "" || content.Quantity == "0" || seen[content.SourceItemNo] || len(*out) >= 100 {
		return imp.ErrInput
	}
	seen[content.SourceItemNo] = true
	*out = append(*out, run.Row{Fields: imp.Fields{Name: content.Name, Type: content.Type, Quantity: content.Quantity, Unit: content.Unit}, SourceItemNo: content.SourceItemNo, SourcePages: location})
	return nil
}
