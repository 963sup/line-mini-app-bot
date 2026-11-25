package summarization

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	summaryApplicationPortSummarization "github.com/963sup/line-mini-app-bot/internal/business/summary/application/port/summarization"
	summaryDomainEntityMessagecollection "github.com/963sup/line-mini-app-bot/internal/business/summary/domain/entity/messagecollection"
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
		input, sources := modelInput(messages)
		if input == "" {
			j.Result = "最近 24 小時沒有可摘要的群組文字。"
		} else {
			byID := map[string]summaryDomainEntityMessagecollection.Message{}
			for _, m := range messages {
				byID[m.ID] = m
			}
			for _, id := range sources {
				if byID[id].ExpireAt.Before(j.ExpireAt) {
					j.ExpireAt = byID[id].ExpireAt
				}
			}
			j.Result, err = s.Model.Summarize(ctx, input)
			if err != nil {
				return err
			}
			j.Result, err = validateSummary(j.Result, len(sources))
			if err != nil {
				return err
			}
			j.Result = fmt.Sprintf("摘要｜最近24小時，採用%d則（最多100則）\n%s\n\n僅整理群聊說法，未做事實查核。", len(sources), j.Result)
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

// modelInput aliases users without losing the source-message association held by the job.
// The byte budget also bounds the maximum possible input token count.
func modelInput(messages []summaryDomainEntityMessagecollection.Message) (string, []string) {
	people := map[string]int{}
	lines := make([]string, 0, len(messages))
	sources := make([]string, 0, len(messages))
	budget := 24000
	for _, m := range messages {
		if len(sources) == 100 {
			break
		}
		if m.Deleted || m.Text == "" {
			continue
		}
		if _, ok := people[m.UserID]; !ok {
			people[m.UserID] = len(people) + 1
		}
		b, _ := json.Marshal(struct{ Source, Speaker, Time, Text string }{fmt.Sprintf("S%d", len(sources)+1), fmt.Sprintf("成員%d", people[m.UserID]), m.OccurredAt.UTC().Format(time.RFC3339), m.Text})
		if len(b) > budget {
			break
		}
		budget -= len(b)
		lines = append(lines, string(b))
		sources = append(sources, m.ID)
	}
	return strings.Join(lines, "\n"), sources
}
