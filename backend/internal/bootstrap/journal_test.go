package bootstrap

import (
	"context"
	"encoding/json"
	"errors"
	"os"
	"strings"
	"testing"
	"time"

	"cloud.google.com/go/firestore"
	teamdb "github.com/963sup/line-mini-app-bot/internal/business/team/adapter/firestore"

	journaldto "github.com/963sup/line-mini-app-bot/internal/business/journal/application/dto"
	journalentity "github.com/963sup/line-mini-app-bot/internal/business/journal/domain/entity"
	project "github.com/963sup/line-mini-app-bot/internal/business/project/application"
	repository "github.com/963sup/line-mini-app-bot/internal/business/repository/application"
	groupdb "github.com/963sup/line-mini-app-bot/internal/linegroup/adapter/firestore"
	groupdto "github.com/963sup/line-mini-app-bot/internal/linegroup/application/dto"
	groupentity "github.com/963sup/line-mini-app-bot/internal/linegroup/domain/entity"
)

type journalGroups struct {
	cursors []string
	err     error
}

func (g *journalGroups) Get(context.Context, string, string) (groupentity.Group, error) {
	return groupentity.Group{}, groupentity.ErrNotFound
}

func (g *journalGroups) List(_ context.Context, actor, cursor string) (groupdto.Page, error) {
	if actor != "reader" {
		return groupdto.Page{}, groupentity.ErrNotFound
	}
	g.cursors = append(g.cursors, cursor)
	if g.err != nil {
		return groupdto.Page{}, g.err
	}
	if cursor == "" {
		return groupdto.Page{NextCursor: "next"}, nil
	}
	return groupdto.Page{}, nil
}

func TestJournalCatalogWiring(t *testing.T) {
	for _, date := range []string{"", "2026-11-22"} {
		t.Run("workDate="+date, func(t *testing.T) {
			groups := &journalGroups{}
			service := newJournals(groups, &groupdb.Store{}, &repository.Service{}, &project.Service{})
			page, err := service.List(t.Context(), "reader", journaldto.Filter{WorkDate: date})
			if err != nil || len(page.Items) != 0 || page.NextCursor != "" {
				t.Fatalf("empty authorized catalog: %+v, %v", page, err)
			}
			if len(groups.cursors) != 2 || groups.cursors[0] != "" || groups.cursors[1] != "next" {
				t.Fatalf("catalog pagination: %v", groups.cursors)
			}
			unavailable := errors.New("catalog unavailable")
			groups.err = unavailable
			if _, err := service.List(t.Context(), "reader", journaldto.Filter{WorkDate: date}); !errors.Is(err, unavailable) {
				t.Fatalf("catalog error: %v", err)
			}
			if _, err := service.List(t.Context(), "reader", journaldto.Filter{GroupID: "Caaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"}); !errors.Is(err, journalentity.ErrNotFound) {
				t.Fatalf("denied group: %v", err)
			}
		})
	}
	if service := newJournals(nil, &groupdb.Store{}, &repository.Service{}, &project.Service{}); service != nil {
		t.Fatal("missing catalog must disable journals")
	}
}

func TestJournalTeamCatalogIntegration(t *testing.T) {
	if os.Getenv("FIRESTORE_EMULATOR_HOST") == "" {
		t.Skip("requires Firestore emulator")
	}
	ctx, cancel := context.WithTimeout(t.Context(), 30*time.Second)
	defer cancel()
	client, err := firestore.NewClient(ctx, "journal-wiring-test")
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = client.Close() }()
	root := client.Collection("channels").NewDoc()
	id := "C" + strings.Repeat("a", 32)
	if _, err := root.Collection("teams").Doc(id).Set(ctx, map[string]any{"name": "Team", "status": "ACTIVE", "membershipMode": "membership"}); err != nil {
		t.Fatal(err)
	}
	if _, err := root.Collection("memberships").NewDoc().Set(ctx, map[string]any{"userId": "reader", "resourceType": "team", "resourceId": id}); err != nil {
		t.Fatal(err)
	}
	allowed := true
	catalog := teamdb.GroupCatalog{Store: teamdb.Store{DB: client, Channel: root.ID, TeamMember: func(*firestore.Transaction, *firestore.DocumentRef, string, string, string, string) (bool, error) {
		return allowed, nil
	}}}
	service := newJournals(catalog, &groupdb.Store{Client: client, ChannelID: root.ID}, &repository.Service{}, &project.Service{})
	date := time.Date(2026, 11, 22, 0, 0, 0, 0, time.UTC)
	entry := journalentity.Entry{ID: strings.Repeat("b", 32), CreatorID: "reader", CreatedAt: date, Draft: journalentity.Draft{
		GroupID: id, WorkDate: date, Today: []journalentity.Work{{Name: "工作", Percent: 40}},
	}}
	payload, err := json.Marshal(entry)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := root.Collection("journalGroups").Doc(id).Collection("entries").Doc(entry.ID).Set(ctx, map[string]any{"payload": string(payload), "order": "2026-11-22/" + entry.ID}); err != nil {
		t.Fatal(err)
	}
	if _, err := root.Collection("journalReceipts").Doc(entry.ID).Set(ctx, map[string]any{"groupId": id}); err != nil {
		t.Fatal(err)
	}
	for _, workDate := range []string{"", "2026-11-22"} {
		page, err := service.List(ctx, "reader", journaldto.Filter{WorkDate: workDate})
		if err != nil || len(page.Items) != 1 || page.Items[0].ID != entry.ID {
			t.Fatalf("workDate %q: %+v, %v", workDate, page, err)
		}
	}
	allowed = false
	page, err := service.List(ctx, "reader", journaldto.Filter{})
	if err != nil || len(page.Items) != 0 {
		t.Fatalf("revoked membership: %+v, %v", page, err)
	}
	if _, err := service.Get(ctx, "reader", entry.ID); !errors.Is(err, journalentity.ErrNotFound) {
		t.Fatalf("revoked detail: %v", err)
	}
}
