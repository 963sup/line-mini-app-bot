package entity

import (
	repositoryDomainEntity "github.com/963sup/line-mini-app-bot/internal/business/repository/domain/entity"
	repositoryDomainEntityLabels "github.com/963sup/line-mini-app-bot/internal/business/repository/domain/entity/labels"
)

import (
	"errors"
	"slices"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/963sup/line-mini-app-bot/internal/business/issue/domain/policy"
)

var (
	ErrInput            = errors.New("invalid issue input")
	ErrNotFound         = errors.New("issue not found")
	ErrForbidden        = errors.New("only the creator may update this issue")
	ErrConflict         = errors.New("issue version conflict")
	ErrCreationConflict = errors.New("issue creation key conflict")
)

type Issue struct {
	MilestoneID                                     string
	ID, RepositoryID, Title, Body, State, CreatorID string
	Number, Version                                 int64
	CreatedAt, UpdatedAt                            time.Time
	LabelIDs                                        []string
}

func (v Issue) CanEdit(actor string) bool { return actor != "" && actor == v.CreatorID }

type (
	Content struct{ Title, Body string }
	Update  struct {
		Content
		State           string
		ExpectedVersion int64
		LabelIDs        *[]string
		MilestoneID     *string
	}
)

func Normalize(c Content) (Content, error) {
	c.Title, c.Body = strings.TrimSpace(c.Title), strings.TrimSpace(c.Body)
	if !utf8.ValidString(c.Title) || !utf8.ValidString(c.Body) || c.Title == "" || utf8.RuneCountInString(c.Title) > 120 || utf8.RuneCountInString(c.Body) > 8000 {
		return Content{}, ErrInput
	}
	return c, nil
}

// ApplyUpdate runs inside the storage transaction against the current item.
func ApplyUpdate(current Issue, actor string, input Update) (Issue, error) {
	if actor == "" || (!current.CanEdit(actor) && (input.Title != current.Title || input.Body != current.Body || input.State != current.State || (input.LabelIDs == nil && input.MilestoneID == nil))) {
		return Issue{}, ErrForbidden
	}
	content, err := Normalize(input.Content)
	if err != nil || !policy.ValidState(input.State) || input.ExpectedVersion < 1 {
		return Issue{}, ErrInput
	}
	if current.Version != input.ExpectedVersion {
		return Issue{}, ErrConflict
	}
	allowed := policy.CanTransition(current.State, input.State)
	if !allowed {
		return Issue{}, ErrInput
	}
	ids := current.LabelIDs
	if input.LabelIDs != nil {
		ids, err = repositoryDomainEntityLabels.NormalizeIDs(*input.LabelIDs)
		if err != nil {
			return Issue{}, ErrInput
		}
	}
	milestoneID := current.MilestoneID
	if input.MilestoneID != nil {
		milestoneID = *input.MilestoneID
		if milestoneID != "" && !repositoryDomainEntity.ValidID(milestoneID) {
			return Issue{}, ErrInput
		}
	}
	if milestoneID == current.MilestoneID && current.Title == content.Title && current.Body == content.Body && current.State == input.State && slices.Equal(current.LabelIDs, ids) {
		return current, nil
	}
	if current.Version >= repositoryDomainEntity.MaxResourceNumber {
		return Issue{}, ErrConflict
	}
	current.Title, current.Body, current.State = content.Title, content.Body, input.State
	current.MilestoneID = milestoneID
	current.LabelIDs = append([]string{}, ids...)
	current.Version++
	return current, nil
}
