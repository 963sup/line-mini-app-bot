package query

import (
	"context"
	"fmt"
	"path/filepath"
	"strings"
	"unicode/utf8"

	"github.com/963sup/line-mini-app-bot/internal/business/object/application/dto"
	"github.com/963sup/line-mini-app-bot/internal/business/object/application/port"
	"github.com/963sup/line-mini-app-bot/internal/business/object/domain/entity"
	"github.com/963sup/line-mini-app-bot/internal/business/object/domain/service"
	"github.com/963sup/line-mini-app-bot/internal/business/object/domain/template"
)

func PrepareTemplateDraft(ctx context.Context, preparer port.TemplatePreparer, actor string, in port.DraftPreparation) (dto.TemplateDraft, error) {
	if actor == "" {
		return dto.TemplateDraft{}, entity.ErrInput
	}
	switch in.Source {
	case "document":
		if in.Filename == "" || len(in.Filename) > 255 || len(in.Data) == 0 || len(in.Data) > 8<<20 || in.Name != "" || in.Conditions != "" {
			return dto.TemplateDraft{}, entity.ErrInput
		}
	case "procedure":
		in.Name = strings.TrimSpace(in.Name)
		if in.Name == "" || !utf8.ValidString(in.Name) || utf8.RuneCountInString(in.Name) > 120 || utf8.RuneCountInString(in.Conditions) > 4000 || len(in.Data) != 0 || in.Filename != "" {
			return dto.TemplateDraft{}, entity.ErrInput
		}
	default:
		return dto.TemplateDraft{}, entity.ErrInput
	}
	if preparer == nil {
		return dto.TemplateDraft{}, entity.ErrUnavailable
	}
	candidates, err := preparer.PrepareTemplate(ctx, actor, in)
	if err != nil {
		return dto.TemplateDraft{}, err
	}
	if len(candidates) == 0 {
		return dto.TemplateDraft{}, port.ErrInvalidCandidate
	}
	name := in.Name
	if in.Source == "document" {
		name = strings.TrimSpace(strings.TrimSuffix(filepath.Base(in.Filename), filepath.Ext(in.Filename)))
	}
	if name == "" {
		name = "文件範本"
	}
	result := dto.TemplateDraft{Draft: dto.TemplateCreation{Name: name, Nodes: []template.Node{{ID: "root", Name: name}}}, Warnings: []string{}, Candidates: candidates, CandidateIndexByNode: make(map[string]int, len(candidates))}
	headings := map[string]string{}
	for i, c := range candidates {
		parent := "root"
		for _, heading := range c.CategoryPath {
			heading = strings.TrimSpace(heading)
			if heading == "" {
				continue
			}
			key := parent + "\x00" + heading
			id, ok := headings[key]
			if !ok {
				id = fmt.Sprintf("n%d", len(result.Draft.Nodes))
				headings[key] = id
				result.Draft.Nodes = append(result.Draft.Nodes, template.Node{ID: id, ParentID: parent, Name: heading})
			}
			parent = id
		}
		unit := ""
		if c.Unit != nil {
			unit = *c.Unit
		}
		id := fmt.Sprintf("n%d", len(result.Draft.Nodes))
		result.CandidateIndexByNode[id] = i
		result.Draft.Nodes = append(result.Draft.Nodes, template.Node{ID: id, ParentID: parent, Name: c.Name, Unit: unit})
	}
	if in.Source == "procedure" {
		result.Warnings = append(result.Warnings, "工序為推導建議，未經施工規範驗證；請核對工法、順序與適用條件。節點順序不代表可執行的前置依賴。")
	}
	if _, err := service.NormalizeTemplate(service.Definition{Name: result.Draft.Name, Nodes: result.Draft.Nodes}); err != nil {
		result.Warnings = append(result.Warnings, "草稿尚不符合範本限制，請核對名稱、單位及最多 100 個節點後再儲存。")
	}
	return result, nil
}
