// Package summary contains the bounded chat-summary contract.
package summary

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"strings"
	"time"
)

type Message struct {
	ID, EventID, UserID, Text string
	OccurredAt, ExpireAt      time.Time
	Epoch                     int64
	Deleted                   bool
}

type Job struct {
	ID, GroupID, Kind, Result, RetryKey, Status string
	Epoch, Revision                             int64
	Cutoff, ExpireAt, LeaseUntil                time.Time
	Lease                                       string
	// Sources contain identifiers only; no additional copy of chat content.
	Sources []string
}

type Store interface {
	Collect(context.Context, string, Message) error
	Unsend(context.Context, string, string) error
	Command(context.Context, string, string, string, int64) (Job, error)
	Claim(context.Context, string, string) (Job, bool, error)
	Messages(context.Context, Job) ([]Message, error)
	SaveResult(context.Context, Job, string, []string) error
	Current(context.Context, Job) (bool, error)
	Done(context.Context, Job) error
	Purge(context.Context, Job) error
}

type Model interface {
	Summarize(context.Context, []Message) (string, error)
}
type Queue interface {
	Enqueue(context.Context, Job) error
}
type Sender interface {
	Send(context.Context, string, string, string) error
}

// RandomID supplies a stable persisted LINE retry key or a unique claim token.
func RandomID() string {
	var b [16]byte
	if _, err := rand.Read(b[:]); err != nil {
		panic("secure random source unavailable")
	}
	b[6] = (b[6] & 15) | 64
	b[8] = (b[8] & 63) | 128
	s := hex.EncodeToString(b[:])
	return s[:8] + "-" + s[8:12] + "-" + s[12:16] + "-" + s[16:20] + "-" + s[20:]
}

// Input aliases users without losing the source-message association held by the job.
// The byte budget also bounds the maximum possible input token count.
func Input(messages []Message) (string, []string) {
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

type Service struct {
	Store  Store
	Model  Model
	Sender Sender
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
		input, sources := Input(messages)
		if input == "" {
			j.Result = "最近 24 小時沒有可摘要的群組文字。"
		} else {
			// Use exactly the bounded set selected by Input, preserving citation numbers.
			selected := make([]Message, 0, len(sources))
			byID := map[string]Message{}
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
