package workbook

import (
	"encoding/csv"
	"io"
	"regexp"
	"strconv"
	"strings"
	"unicode/utf8"

	imp "github.com/963sup/line-mini-app-bot/internal/business/objectimport/domain/entity"
	run "github.com/963sup/line-mini-app-bot/internal/business/objectimport/domain/entity/automatic"
)

var divider = regexp.MustCompile("^:?-{3,}:?$")

// ReadText requires an explicit header. It never guesses quantities from prose.
func ReadText(format string, data []byte) ([]run.Row, error) {
	if !utf8.Valid(data) || len(data) > 300<<10 || strings.IndexByte(string(data), 0) >= 0 {
		return nil, imp.ErrInput
	}
	text := strings.TrimPrefix(string(data), "\uFEFF")
	var header []string
	var cols map[string]int
	var out []run.Row
	seen := map[string]bool{}
	read := func(fields []string, line int) error {
		if header == nil {
			header = fields
			var err error
			cols, err = columns(header)
			return err
		}
		if len(fields) != len(header) {
			return imp.ErrInput
		}
		selected := map[string]string{}
		for field, i := range cols {
			selected[field] = strings.TrimSpace(fields[i])
		}
		return appendRow(&out, seen, selected, "L"+strconv.Itoa(line))
	}
	switch format {
	case "md":
		separator := false
		for i, line := range strings.Split(text, "\n") {
			line = strings.TrimSpace(line)
			if line == "" {
				continue
			}
			if !strings.Contains(line, "|") || strings.Contains(line, `\|`) {
				return nil, imp.ErrInput
			}
			fields := strings.Split(strings.TrimSuffix(strings.TrimPrefix(line, "|"), "|"), "|")
			if header != nil && !separator {
				if len(fields) != len(header) {
					return nil, imp.ErrInput
				}
				for _, cell := range fields {
					if !divider.MatchString(strings.TrimSpace(cell)) {
						return nil, imp.ErrInput
					}
				}
				separator = true
				continue
			}
			if err := read(fields, i+1); err != nil {
				return nil, err
			}
		}
	case "csv", "tsv", "txt":
		r := csv.NewReader(strings.NewReader(text))
		if format != "csv" {
			r.Comma = '\t'
		}
		for {
			fields, err := r.Read()
			if err == io.EOF {
				break
			}
			if err != nil {
				return nil, imp.ErrInput
			}
			line, _ := r.FieldPos(0)
			if err = read(fields, line); err != nil {
				return nil, err
			}
		}
	default:
		return nil, imp.ErrInput
	}
	if len(out) == 0 {
		return nil, imp.ErrInput
	}
	return out, nil
}
