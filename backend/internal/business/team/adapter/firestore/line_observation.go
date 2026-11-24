package teamdb

import (
	"context"
	"time"

	"cloud.google.com/go/firestore"
	groupport "github.com/963sup/line-mini-app-bot/internal/business/team/application/port"
	groupentity "github.com/963sup/line-mini-app-bot/internal/business/team/domain/entity"
	groupevent "github.com/963sup/line-mini-app-bot/internal/business/team/domain/event"
	"github.com/963sup/line-mini-app-bot/internal/platform/line/lineid"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// LINEObservation keeps the LINE event ordering on the Team document.
type LINEObservation struct {
	Store
	Membership groupport.Membership
}

func (s LINEObservation) ref(id string) *firestore.DocumentRef {
	return s.DB.Collection("channels").Doc(s.Channel).Collection("teams").Doc(id)
}

func (s LINEObservation) GetByID(ctx context.Context, id string) (groupentity.Group, error) {
	snap, err := s.ref(id).Get(ctx)
	if status.Code(err) == codes.NotFound {
		return groupentity.Group{}, groupentity.ErrNotFound
	}
	if err != nil {
		return groupentity.Group{}, err
	}
	var team struct {
		Name   string `firestore:"name"`
		Status string `firestore:"status"`
	}
	if err := snap.DataTo(&team); err != nil {
		return groupentity.Group{}, err
	}
	return groupentity.Group{ID: id, GroupID: id, Name: team.Name, Status: team.Status}, nil
}

// Get supports legacy team visibility using current LINE membership and Team state.
func (s LINEObservation) Get(ctx context.Context, actor, id string) (groupentity.Group, error) {
	if !lineid.ValidGroup(id) || actor == "" {
		return groupentity.Group{}, groupentity.ErrNotFound
	}
	group, err := s.GetByID(ctx, id)
	if err != nil {
		return groupentity.Group{}, err
	}
	if group.Status != "ACTIVE" {
		return groupentity.Group{}, groupentity.ErrNotFound
	}
	if s.Membership == nil {
		return groupentity.Group{}, groupport.ErrUnavailable
	}
	ok, err := s.Membership.IsMember(ctx, id, actor)
	if err != nil {
		return groupentity.Group{}, err
	}
	if !ok {
		return groupentity.Group{}, groupentity.ErrNotFound
	}
	return group, nil
}

func (s LINEObservation) Observe(ctx context.Context, event groupevent.Observation) error {
	if !lineid.ValidGroup(event.GroupID) || event.EventID == "" || event.At.IsZero() ||
		(event.Status != "ACTIVE" && event.Status != "INACTIVE") {
		return groupport.ErrUnavailable
	}
	ref := s.ref(event.GroupID)
	return s.DB.RunTransaction(ctx, func(_ context.Context, tx *firestore.Transaction) error {
		snap, err := tx.Get(ref)
		if err == nil {
			var previous struct {
				At     time.Time `firestore:"lineEventAt"`
				ID     string    `firestore:"lineEventId"`
				Status string    `firestore:"status"`
			}
			if err := snap.DataTo(&previous); err != nil {
				return err
			}
			if previous.At.IsZero() != (previous.ID == "") {
				return groupport.ErrUnavailable
			}
			if previous.At.IsZero() {
				// Old projections have no LINE event watermark. Their last
				// server write is a conservative cutoff for delayed events.
				if !event.At.After(snap.UpdateTime) {
					return nil
				}
			}
			if event.At.Before(previous.At) || event.At.Equal(previous.At) &&
				(event.EventID == previous.ID || previous.Status == "INACTIVE") {
				return nil
			}
		} else if status.Code(err) != codes.NotFound {
			return err
		}
		fields := map[string]any{"lineEventAt": event.At, "lineEventId": event.EventID, "status": event.Status}
		if event.Name != "" {
			fields["name"] = event.Name
		}
		return tx.Set(ref, fields, firestore.MergeAll)
	})
}
