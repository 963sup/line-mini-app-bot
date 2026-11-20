package workbook

import (
	"bytes"
	"context"
	"encoding/xml"
	"io"
	"math"
	"os/exec"
	"regexp"
	"slices"
	"sort"
	"strings"
	"time"

	imp "github.com/963sup/line-mini-app-bot/internal/business/objectimport/domain/entity"
)

type pdfWord struct {
	X    float64 `xml:"xMin,attr"`
	Y    float64 `xml:"yMin,attr"`
	Text string  `xml:",chardata"`
}

var itemID = regexp.MustCompile(`^[A-Za-z0-9][A-Za-z0-9._/-]*$`)

// ReadPDF bounds conversion time, pages and extracted bytes before decoding.
func ReadPDF(ctx context.Context, file []byte) (imp.PDFInput, error) {
	if !bytes.HasPrefix(file, []byte("%PDF-")) {
		return imp.PDFInput{}, imp.ErrInput
	}
	ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()
	cmd := exec.CommandContext(ctx, "pdftotext", "-bbox", "-l", "51", "-", "-")
	cmd.Stdin = bytes.NewReader(file)
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return imp.PDFInput{}, err
	}
	if err = cmd.Start(); err != nil {
		return imp.PDFInput{}, err
	}
	const limit = 8 << 20
	data, err := io.ReadAll(io.LimitReader(stdout, limit+1))
	stdout.Close()
	if wait := cmd.Wait(); err != nil || wait != nil || len(data) > limit {
		return imp.PDFInput{}, imp.ErrInput
	}
	var doc struct {
		Pages []struct {
			Words []pdfWord `xml:"word"`
		} `xml:"body>doc>page"`
	}
	if xml.Unmarshal(data, &doc) != nil || len(doc.Pages) > 50 {
		return imp.PDFInput{}, imp.ErrInput
	}
	var pages []imp.Page
	var located []imp.PDFItem
	var columns []pdfWord
	var lastItem string
	var fullText []imp.Page
	layoutVerified := true
	for i, page := range doc.Pages {
		var raw strings.Builder
		for _, word := range page.Words {
			if raw.Len() > 0 {
				raw.WriteByte(' ')
			}
			raw.WriteString(word.Text)
		}
		text := strings.TrimSpace(raw.String())
		if text != "" {
			fullText = append(fullText, imp.Page{Number: i + 1, Text: pageText(page.Words)})
		}
		if !layoutVerified {
			continue
		}
		detail, items, err := detailText(page.Words, &columns, &lastItem)
		if err == nil && detail != "" {
			pages = append(pages, imp.Page{Number: i + 1, Text: detail})
			located = append(located, items...)
			continue
		}
		if err == nil && i == 0 && (strings.Contains(text, "封面") || strings.Contains(strings.ToLower(text), "cover")) {
			continue
		}
		layoutVerified = false
	}
	if !layoutVerified && len(fullText) > 0 {
		return imp.PDFInput{Pages: fullText, Evidence: imp.PDFPageText{}}, nil
	}
	if len(pages) == 0 {
		return imp.PDFInput{}, imp.ErrInput
	}
	return imp.PDFInput{Pages: pages, Evidence: imp.LocatedPDFItems{Items: located}}, nil
}

// Poppler returns words in reading order. Preserve line breaks inside each page.
func pageText(words []pdfWord) string {
	var text strings.Builder
	for i, word := range words {
		if i > 0 {
			if math.Abs(word.Y-words[i-1].Y) >= 3 || word.X < words[i-1].X {
				text.WriteByte('\n')
			} else {
				text.WriteByte(' ')
			}
		}
		text.WriteString(word.Text)
	}
	return strings.TrimSpace(text.String())
}

func detailText(words []pdfWord, previous *[]pdfWord, lastItem *string) (string, []imp.PDFItem, error) {
	headers := [][]string{{"項次", "項目"}, {"料號/品名", "品名", "名稱", "項目名稱"}, {"數量", "契約數量"}, {"單位"}, {"單價", "契約單價", "價格"}}
	var cols []pdfWord
	for i, labels := range headers {
		for _, w := range words {
			if slices.Contains(labels, w.Text) && (len(cols) == 0 || (w.X > cols[len(cols)-1].X && math.Abs(w.Y-cols[0].Y) <= 3)) {
				cols = append(cols, w)
				break
			}
		}
		if len(cols) != i+1 {
			break
		}
	}
	continuation := len(cols) == 0 && len(*previous) > 0 && *lastItem != ""
	if continuation {
		cols = slices.Clone(*previous)
		cols[0].Y = -6
	} else if len(cols) == 0 {
		return "", nil, nil
	} else if len(cols) < 4 {
		return "", nil, imp.ErrInput
	} else {
		*previous = cols
	}
	right := cols[3].X + (cols[3].X-cols[2].X)/2
	if len(cols) == 4 {
		for _, word := range words {
			if word.X > right && math.Abs(word.Y-cols[0].Y) <= 3 {
				return "", nil, imp.ErrInput
			}
		}
	}
	if len(cols) == 5 {
		right = (cols[3].X + cols[4].X) / 2
	}
	bottom := math.Inf(1)
	for _, w := range words {
		switch w.Text {
		case "Page", "未稅總計", "含稅總價", "採購安全衛生審核須知":
			if w.Y > cols[0].Y && w.Y < bottom {
				bottom = w.Y
			}
		}
	}
	if math.IsInf(bottom, 1) {
		return "", nil, imp.ErrInput
	}
	var cells []pdfWord
	for _, w := range words {
		if w.Y > cols[0].Y+5 && w.Y < bottom-3 && w.X >= cols[0].X-2 && w.X < right {
			cells = append(cells, w)
		}
	}
	sort.SliceStable(cells, func(i, j int) bool { return cells[i].Y < cells[j].Y })
	var lines []string
	items := []imp.PDFItem{}
	for len(cells) > 0 {
		n := 1
		for n < len(cells) && cells[n].Y-cells[0].Y < 3 {
			n++
		}
		line := cells[:n]
		cells = cells[n:]
		sort.Slice(line, func(i, j int) bool { return line[i].X < line[j].X })
		parts := make([]string, 0, len(line))
		for _, w := range line {
			parts = append(parts, w.Text)
		}
		value := strings.Join(parts, " ")
		if strings.Contains(value, "[[item:") {
			return "", nil, imp.ErrInput
		}
		if strings.HasPrefix(value, "(Cost Ref:") || strings.HasPrefix(value, "(Customer PO:") {
			continue
		}
		if continuation && len(lines) == 0 && line[0].X < cols[1].X-2 {
			return "", nil, imp.ErrInput
		}
		if line[0].X < cols[1].X-2 {
			if !itemID.MatchString(parts[0]) {
				return "", nil, imp.ErrInput
			}
			item := imp.PDFItem{ItemNo: parts[0]}
			for _, word := range line[1:] {
				switch {
				case word.X < cols[2].X-2:
					item.Name = strings.TrimSpace(item.Name + " " + word.Text)
				case word.X < cols[3].X-2:
					item.Quantity = strings.TrimSpace(item.Quantity + " " + word.Text)
				default:
					item.Unit = strings.TrimSpace(item.Unit + " " + word.Text)
				}
			}
			items = append(items, item)
			*lastItem = item.ItemNo
			value = "[[item:" + parts[0] + "]] " + value
		}
		lines = append(lines, value)
	}
	if len(lines) == 0 {
		return "", nil, imp.ErrInput
	}
	return strings.Join(lines, "\n"), items, nil
}
