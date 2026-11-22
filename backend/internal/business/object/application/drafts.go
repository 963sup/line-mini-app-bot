package application

import (
	"context"
	"fmt"
	"path/filepath"
	"strings"
	"unicode/utf8"

	"github.com/963sup/line-mini-app-bot/internal/business/object/application/dto"
	"github.com/963sup/line-mini-app-bot/internal/business/object/domain/entity"
	"github.com/963sup/line-mini-app-bot/internal/business/object/domain/template"
	operations "github.com/963sup/line-mini-app-bot/internal/business/operations/application"
)

// DraftPreparation is a read-only operation. Identity comes from the authenticated request.
type DraftPreparation struct {
	Source, Filename, Name, Conditions string
	Data                               []byte
}

type TemplateDraft struct {
	Draft      dto.TemplateCreation
	Warnings   []string
	Candidates []operations.Candidate
}

// TemplatePreparer is owned by Object; bootstrap adapts the preparation workflow.
type TemplatePreparer interface {
	PrepareTemplate(context.Context, string, DraftPreparation) ([]operations.Candidate, error)
}

func (s *Service) PrepareTemplateDraft(ctx context.Context, actor string, in DraftPreparation) (TemplateDraft, error) {
	if actor == "" {
		return TemplateDraft{}, entity.ErrInput
	}
	switch in.Source {
	case "document":
		if in.Filename == "" || len(in.Filename) > 255 || len(in.Data) == 0 || len(in.Data) > 8<<20 || in.Name != "" || in.Conditions != "" {
			return TemplateDraft{}, entity.ErrInput
		}
	case "procedure":
		in.Name = strings.TrimSpace(in.Name)
		if in.Name == "" || !utf8.ValidString(in.Name) || utf8.RuneCountInString(in.Name) > 120 || utf8.RuneCountInString(in.Conditions) > 4000 || len(in.Data) != 0 || in.Filename != "" {
			return TemplateDraft{}, entity.ErrInput
		}
	default:
		return TemplateDraft{}, entity.ErrInput
	}
	if s == nil || s.TemplatePreparer == nil {
		return TemplateDraft{}, entity.ErrUnavailable
	}
	candidates, err := s.TemplatePreparer.PrepareTemplate(ctx, actor, in)
	if err != nil {
		return TemplateDraft{}, err
	}
	if len(candidates) == 0 {
		return TemplateDraft{}, operations.ErrInvalidOutput
	}
	name := in.Name
	if in.Source == "document" {
		name = strings.TrimSpace(strings.TrimSuffix(filepath.Base(in.Filename), filepath.Ext(in.Filename)))
	}
	if name == "" {
		name = "文件範本"
	}
	result := TemplateDraft{Draft: dto.TemplateCreation{Name: name, Nodes: []template.Node{{ID: "root", Name: name}}}, Warnings: []string{}, Candidates: candidates}
	headings := map[string]string{}
	for _, c := range candidates {
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
		result.Draft.Nodes = append(result.Draft.Nodes, template.Node{ID: fmt.Sprintf("n%d", len(result.Draft.Nodes)), ParentID: parent, Name: c.Name, Unit: unit})
	}
	if in.Source == "procedure" {
		result.Warnings = append(result.Warnings, "工序為推導建議，未經施工規範驗證；請核對工法、順序與適用條件。節點順序不代表可執行的前置依賴。")
	}
	if _, err := NormalizeTemplate(result.Draft); err != nil {
		result.Warnings = append(result.Warnings, "草稿尚不符合範本限制，請核對名稱、單位及最多 100 個節點後再儲存。")
	}
	return result, nil
}
