package firestore

import (
	"context"

	db "cloud.google.com/go/firestore"
	"github.com/963sup/line-mini-app-bot/internal/poll"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type PollStore struct {
	Client    *db.Client
	ChannelID string
}

func (s *PollStore) ref(id string) *db.DocumentRef {
	return s.Client.Collection("channels").Doc(s.ChannelID).Collection("polls").Doc(id)
}

func (s *PollStore) Create(ctx context.Context, group, user, event string, p poll.Poll) (poll.Result, error) {
	ref := s.ref(event)
	p.ID, p.GroupID, p.CreatorID = event, group, user
	err := s.Client.RunTransaction(ctx, func(ctx context.Context, tx *db.Transaction) error {
		snap, err := tx.Get(ref)
		if err == nil {
			var old poll.Poll
			if err := snap.DataTo(&old); err != nil {
				return err
			}
			if old.GroupID != group || old.CreatorID != user {
				return poll.ErrInvalid
			}
			p = old
			return nil
		}
		if status.Code(err) != codes.NotFound {
			return err
		}
		return tx.Create(ref, p)
	})
	return poll.Result{Poll: p, Code: "created"}, err
}

func (s *PollStore) Vote(ctx context.Context, group, user, event, id string, option int) (poll.Result, error) {
	ref := s.ref(id)
	vote := ref.Collection("votes").Doc(documentKey(user))
	receipt := ref.Collection("events").Doc(documentKey(event))
	result := poll.Result{}
	err := s.Client.RunTransaction(ctx, func(ctx context.Context, tx *db.Transaction) error {
		result = poll.Result{}
		snap, err := tx.Get(receipt)
		if err == nil {
			return snap.DataTo(&result)
		}
		if status.Code(err) != codes.NotFound {
			return err
		}
		snap, err = tx.Get(ref)
		if err != nil {
			return err
		}
		if err := snap.DataTo(&result.Poll); err != nil {
			return err
		}
		if result.Poll.Closed || result.Poll.GroupID != group || option < 0 || option >= len(result.Poll.Options) || len(result.Poll.Counts) != len(result.Poll.Options) {
			return poll.ErrInvalid
		}
		old, err := tx.Get(vote)
		previous := -1
		if err == nil {
			var data struct {
				Option int `firestore:"option"`
			}
			if err := old.DataTo(&data); err != nil {
				return err
			}
			previous = data.Option
		} else if status.Code(err) != codes.NotFound {
			return err
		}
		if previous >= len(result.Poll.Counts) {
			return poll.ErrInvalid
		}
		result.Code = "voted"
		if previous == option {
			result.Code = "unchanged"
		} else {
			if previous >= 0 {
				result.Poll.Counts[previous]--
				result.Code = "changed"
			}
			result.Poll.Counts[option]++
			if err := tx.Set(vote, map[string]any{"option": option}); err != nil {
				return err
			}
			if err := tx.Update(ref, []db.Update{{Path: "counts", Value: result.Poll.Counts}}); err != nil {
				return err
			}
		}
		return tx.Create(receipt, result)
	})
	return result, err
}

func (s *PollStore) Get(ctx context.Context, group, id string) (poll.Poll, error) {
	var p poll.Poll
	snap, err := s.ref(id).Get(ctx)
	if err != nil {
		return p, err
	}
	if err := snap.DataTo(&p); err != nil {
		return p, err
	}
	if p.GroupID != group {
		return poll.Poll{}, poll.ErrInvalid
	}
	return p, nil
}

func (s *PollStore) Close(ctx context.Context, group, user, id string) (poll.Poll, error) {
	var p poll.Poll
	err := s.Client.RunTransaction(ctx, func(ctx context.Context, tx *db.Transaction) error {
		snap, err := tx.Get(s.ref(id))
		if err != nil {
			return err
		}
		if err := snap.DataTo(&p); err != nil {
			return err
		}
		if p.GroupID != group || p.CreatorID != user {
			return poll.ErrInvalid
		}
		p.Closed = true
		return tx.Update(s.ref(id), []db.Update{{Path: "closed", Value: true}})
	})
	return p, err
}
