package summarization

import (
	"context"
	"fmt"
	summaryApplicationPortSummarization "github.com/963sup/line-mini-app-bot/internal/business/summary/application/port/summarization"
	summaryDomainEntityMessagecollection "github.com/963sup/line-mini-app-bot/internal/business/summary/domain/entity/messagecollection"
	summaryDomainEntitySummarization "github.com/963sup/line-mini-app-bot/internal/business/summary/domain/entity/summarization"
)

type Service struct {
	Store  summaryApplicationPortSummarization.Store
	Model  summaryApplicationPortSummarization.Model
	Sender summaryApplicationPortSummarization.Sender
}

func (s *Service) Run(ctx context.Context, group, id string) error {
	j, claimed, err := s.Store.Claim(ctx, group, id)
	if err != nil || !claimed {
		return err
	}
	if j.Kind == "purge" {
		if err := s.Store.Purge(ctx, j); err != nil {
			return err
		}
		return s.Store.Done(ctx, j)
	}
	current, err := s.Store.Current(ctx, j)
	if err != nil {
		return err
	}
	if !current {
		return s.Store.Done(ctx, j)
	}
	if j.Result == "" {
		messages, err := s.Store.Messages(ctx, j)
		if err != nil {
			return err
		}
		input, sources := summaryDomainEntitySummarization.Input(messages)
		if input == "" {
			j.Result = "最近 24 小時沒有可摘要的群組文字。"
		} else {
			// Use exactly the bounded set selected by Input, preserving citation numbers.
			selected := make([]summaryDomainEntityMessagecollection.Message, 0, len(sources))
			byID := map[string]summaryDomainEntityMessagecollection.Message{}
			for _, m := range messages {
				byID[m.ID] = m
			}
			for _, id := range sources {
				selected = append(selected, byID[id])
				if byID[id].ExpireAt.Before(j.ExpireAt) {
					j.ExpireAt = byID[id].ExpireAt
				}
			}
			j.Result, err = s.Model.Summarize(ctx, selected)
			if err != nil {
				return err
			}
			j.Result, err = validateSummary(j.Result, len(sources))
			if err != nil {
				return err
			}
			j.Result = fmt.Sprintf("群聊摘要｜最近24小時，採用%d則（最多100則）\n%s\n\n僅整理群聊說法，未做事實查核。", len(sources), j.Result)
		}
		if err := s.Store.SaveResult(ctx, j, j.Result, sources); err != nil {
			return err
		}
	}
	current, err = s.Store.Current(ctx, j)
	if err != nil {
		return err
	}
	if !current {
		return s.Store.Done(ctx, j)
	}
	if err := s.Sender.Send(ctx, j.GroupID, j.Result, j.RetryKey); err != nil {
		return err
	}
	return s.Store.Done(ctx, j)
}
