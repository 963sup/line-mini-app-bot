package wiki

import (
	"context"
	"errors"
	"testing"

	wikiport "github.com/963sup/line-mini-app-bot/internal/business/repository/application/port/wiki"
	"github.com/963sup/line-mini-app-bot/internal/business/repository/domain/entity"
	wikimodel "github.com/963sup/line-mini-app-bot/internal/business/repository/domain/entity/wiki"
)

type writeStore struct {
	wikiport.Store
	state  wikiport.WriteState
	writes int
}

func (s *writeStore) WriteWiki(_ context.Context, _ entity.Repository, _ wikiport.WriteTarget, decide wikiport.DecideWrite) (wikimodel.Page, error) {
	next, changed, err := decide(s.state)
	if err != nil {
		return wikimodel.Page{}, err
	}
	if changed {
		s.writes++
		s.state.Current = &next
		if s.state.FirstRevision == nil {
			s.state.FirstRevision = &next
		}
	}
	return next, nil
}

func TestCreateReplayAfterEditDoesNotAddRevision(t *testing.T) {
	store := &writeStore{state: wikiport.WriteState{ID: "page"}}
	service := Service{Store: store}
	save := func(id string, expected int64, title string) (wikimodel.Page, error) {
		return service.Save(context.Background(), entity.Repository{}, "actor", id, "key", expected, title, "body")
	}
	created, err := save("", 0, " Original ")
	if err != nil || created.ID != "page" || created.Version != 1 || created.Title != "Original" {
		t.Fatalf("create = %+v, %v", created, err)
	}
	updated, err := save(created.ID, 1, "Edited")
	if err != nil || updated.Version != 2 {
		t.Fatalf("update = %+v, %v", updated, err)
	}
	replayed, err := save("", 0, "Original")
	if err != nil || replayed != updated {
		t.Fatalf("replay = %+v, %v, want %+v", replayed, err, updated)
	}
	if _, err := save("", 0, "Edited"); !errors.Is(err, wikimodel.ErrConflict) {
		t.Fatalf("different content with same key: %v", err)
	}
	if _, err := save(created.ID, 1, "Stale"); !errors.Is(err, wikimodel.ErrConflict) {
		t.Fatalf("stale edit: %v", err)
	}
	if _, err := save(created.ID, 2, "Edited"); err != nil || store.writes != 2 {
		t.Fatalf("no-op edit: writes=%d, err=%v", store.writes, err)
	}
}
