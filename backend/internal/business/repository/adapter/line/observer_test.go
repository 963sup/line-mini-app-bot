package line

import (
	repositoryApplicationDto "github.com/963sup/line-mini-app-bot/internal/business/repository/application/dto"
	repositoryDomainEntity "github.com/963sup/line-mini-app-bot/internal/business/repository/domain/entity"
)

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	transport "github.com/963sup/line-mini-app-bot/internal/transport/line"
	"github.com/line/line-bot-sdk-go/v8/linebot/webhook"
)

type observerStore struct {
	name               string
	lookupErr, saveErr error
	lookups            int
	saved              []repositoryDomainEntity.Observation
}

func (s *observerStore) Observe(_ context.Context, o repositoryDomainEntity.Observation) error {
	s.saved = append(s.saved, o)
	return s.saveErr
}

func (s *observerStore) GetByGroupID(context.Context, string) (repositoryDomainEntity.Repository, error) {
	s.lookups++
	return repositoryDomainEntity.Repository{Name: s.name}, s.lookupErr
}

func (*observerStore) GetByID(context.Context, string) (repositoryDomainEntity.Repository, error) {
	return repositoryDomainEntity.Repository{}, repositoryDomainEntity.ErrNotFound
}

func (*observerStore) Candidates(context.Context, string, string, string) (repositoryApplicationDto.Page, error) {
	return repositoryApplicationDto.Page{}, nil
}

func TestObserverPreservesDomainNameLifecycleAndFailures(t *testing.T) {
	gid := "C" + strings.Repeat("a", 32)
	source := webhook.GroupSource{GroupId: gid, UserId: "actor"}
	message := webhook.MessageEvent{Source: source, Timestamp: 1700000000123, Mode: "active", WebhookEventId: "message"}
	join := webhook.JoinEvent{Source: source, Timestamp: message.Timestamp, Mode: "active", WebhookEventId: "join"}
	leave := webhook.LeaveEvent{Source: source, Timestamp: message.Timestamp, Mode: "active", WebhookEventId: "leave"}
	failure := errors.New("domain storage failure")
	for _, tc := range []struct {
		name                  string
		event                 webhook.EventInterface
		current               string
		lookupErr, saveErr    error
		lookups, names, saves int
		state                 string
		failed                bool
	}{
		{"join refreshes name", join, "stored", nil, nil, 0, 1, 1, "ACTIVE", false},
		{"message retains name", message, "stored", nil, nil, 1, 0, 1, "ACTIVE", false},
		{"new group looks up name", message, "", repositoryDomainEntity.ErrNotFound, nil, 1, 1, 1, "ACTIVE", false},
		{"leave never looks up name", leave, "stored", nil, nil, 0, 0, 1, "INACTIVE", false},
		{"lookup failure propagates", message, "", failure, nil, 1, 0, 0, "", true},
		{"save failure propagates", message, "stored", nil, failure, 1, 0, 1, "ACTIVE", true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			store := &observerStore{name: tc.current, lookupErr: tc.lookupErr, saveErr: tc.saveErr}
			names := 0
			o := Observer{Store: store, GroupName: func(_ context.Context, id string) string {
				names++
				if id != gid {
					t.Fatal("wrong group")
				}
				return "current name"
			}}
			group, ok := transport.DecodeGroupEvent(tc.event)
			if !ok {
				t.Fatal("invalid test group event")
			}
			err := o.Observe(context.Background(), tc.event, &group)
			if errors.Is(err, failure) != tc.failed || store.lookups != tc.lookups || names != tc.names || len(store.saved) != tc.saves {
				t.Fatalf("err=%v lookups=%d names=%d saves=%d", err, store.lookups, names, len(store.saved))
			}
			if tc.saves > 0 {
				got := store.saved[0]
				if got.OwnerID != gid || got.Status != tc.state || !got.At.Equal(time.UnixMilli(message.Timestamp)) || len(got.Users) != 1 || got.Users[0] != "actor" {
					t.Fatalf("domain observation changed: %+v", got)
				}
				if (got.Name == "current name") != (tc.names > 0) {
					t.Fatalf("name lookup changed: %+v", got)
				}
			}
		})
	}
}

func TestFollowCreatesPersonalRepositoryObservation(t *testing.T) {
	owner := "U" + strings.Repeat("1", 32)
	store := &observerStore{}
	observer := Observer{Store: store}
	event := webhook.FollowEvent{Source: webhook.UserSource{UserId: owner}, Mode: "active", Timestamp: 123, WebhookEventId: "follow"}
	err := observer.Observe(context.Background(), event, nil)
	if err != nil || len(store.saved) != 1 {
		t.Fatalf("follow: %v %v", store.saved, err)
	}
	saved := store.saved[0]
	if saved.OwnerID != owner || len(saved.Users) != 1 || saved.Users[0] != owner || store.lookups != 0 {
		t.Fatal(saved)
	}
}
