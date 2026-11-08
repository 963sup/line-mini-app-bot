package poll

import (
	"context"
	"errors"
	"strings"
	"unicode/utf8"
)

var ErrInvalid = errors.New("invalid poll")

type Poll struct {
	ID        string   `firestore:"id"`
	Closed    bool     `firestore:"closed"`
	GroupID   string   `firestore:"groupID"`
	CreatorID string   `firestore:"creatorID"`
	Question  string   `firestore:"question"`
	Options   []string `firestore:"options"`
	Counts    []int    `firestore:"counts"`
}

type Result struct {
	Poll Poll
	Code string
}

type Store interface {
	Create(context.Context, string, string, string, Poll) (Result, error)
	Vote(context.Context, string, string, string, string, int) (Result, error)
	Close(context.Context, string, string, string) (Poll, error)
	Get(context.Context, string, string) (Poll, error)
}

func Parse(text string) (Poll, error) {
	parts := strings.Split(strings.TrimSpace(strings.TrimPrefix(text, "投票 ")), "｜")
	if len(parts) < 3 || len(parts) > 5 {
		return Poll{}, ErrInvalid
	}
	p := Poll{Question: strings.TrimSpace(parts[0]), Options: make([]string, 0, len(parts)-1), Counts: make([]int, len(parts)-1)}
	if p.Question == "" || utf8.RuneCountInString(p.Question) > 100 {
		return Poll{}, ErrInvalid
	}
	seen := map[string]bool{}
	for _, raw := range parts[1:] {
		option := strings.TrimSpace(raw)
		if option == "" || utf8.RuneCountInString(option) > 40 || seen[option] {
			return Poll{}, ErrInvalid
		}
		seen[option] = true
		p.Options = append(p.Options, option)
	}
	return p, nil
}
