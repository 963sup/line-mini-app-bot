// Package firestore persists immutable journal entries and their creation receipts.
package firestore

import (
	journalquery "github.com/963sup/line-mini-app-bot/internal/business/journal/application/query"
	repositoryDomainEntity "github.com/963sup/line-mini-app-bot/internal/business/repository/domain/entity"
)

import (
	projectDomainEntity "github.com/963sup/line-mini-app-bot/internal/business/project/domain/entity"
)

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"sort"
	"strings"
	"time"

	db "cloud.google.com/go/firestore"
	groupentity "github.com/963sup/line-mini-app-bot/internal/linegroup/domain/entity"

	journaldto "github.com/963sup/line-mini-app-bot/internal/business/journal/application/dto"
	journalentity "github.com/963sup/line-mini-app-bot/internal/business/journal/domain/entity"

	"github.com/963sup/line-mini-app-bot/internal/platform/firestore/key"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type Store struct {
	Client  *db.Client
	Channel string
	journalquery.ReferenceCatalog
}
type record struct {
	Payload string `firestore:"payload"`
	Order   string `firestore:"order"`
}
type receipt struct {
	Group      string `firestore:"groupId"`
	Repository string `firestore:"repositoryId,omitempty"`
	Hash       string `firestore:"hash"`
}

func (s *Store) root() *db.DocumentRef { return s.Client.Collection("channels").Doc(s.Channel) }
func (s *Store) entries(group string) *db.CollectionRef {
	return s.root().Collection("journalGroups").Doc(group).Collection("entries")
}

func (s *Store) repositoryEntries(id string) *db.CollectionRef {
	return s.root().Collection("journalRepositories").Doc(id).Collection("entries")
}

func missing(e error) error {
	if status.Code(e) == codes.NotFound || errors.Is(e, groupentity.ErrNotFound) || errors.Is(e, repositoryDomainEntity.ErrNotFound) || errors.Is(e, projectDomainEntity.ErrNotFound) {
		return journalentity.ErrNotFound
	}
	return e
}

func (s *Store) group(ctx context.Context, actor, group string) error {
	_, e := s.Groups.Get(ctx, actor, group)
	return missing(e)
}

func (s *Store) reference(ctx context.Context, actor, group string, r journalentity.Reference) error {
	switch r.Kind {
	case "repository":
		owner, e := s.RepositoryOwner(ctx, actor, r.ID)
		if e != nil {
			return missing(e)
		}
		if owner == group {
			return nil
		}
	case "project":
		owner, e := s.ProjectOwner(ctx, actor, group, r.ID)
		if e != nil {
			return missing(e)
		}
		if owner == group {
			return nil
		}
	}
	return journalentity.ErrNotFound
}

func (s *Store) checkRefs(ctx context.Context, actor string, d journalentity.Draft, redact bool, cache map[journalentity.Reference]error) error {
	for _, rows := range [][]journalentity.Work{d.Today, d.Tomorrow} {
		for i := range rows {
			work := &rows[i]
			if work.Reference == nil {
				continue
			}
			r := *work.Reference
			err, ok := cache[r]
			if !ok {
				if d.RepositoryID != "" && r.Kind == "repository" && r.ID != d.RepositoryID {
					err = journalentity.ErrNotFound
				} else {
					err = s.reference(ctx, actor, d.GroupID, r)
				}
				cache[r] = err
			}
			if err := work.ApplyReferenceCheck(err, redact); err != nil {
				return err
			}
		}
	}
	return nil
}

// snapshotWork checks repository-scoped sources and captures names in the publish transaction.
func (s *Store) snapshotWork(tx *db.Transaction, group string, w *journalentity.Work) error {
	if w.ObjectID == "" && len(w.SpaceIDs) == 0 {
		return nil
	}
	if w.Reference == nil || w.Reference.Kind != "repository" || len(w.SpaceIDs) > 20 {
		return journalentity.ErrInput
	}
	repo := s.root().Collection("repositories").Doc(w.Reference.ID)
	snap, err := tx.Get(repo)
	if err != nil {
		return missing(err)
	}
	if snap.Data()["ownerId"] != group || snap.Data()["status"] != "ACTIVE" {
		return journalentity.ErrNotFound
	}
	readName := func(ref *db.DocumentRef) (string, error) {
		snap, err := tx.Get(ref)
		if err != nil {
			return "", missing(err)
		}
		data := snap.Data()
		if data["id"] != ref.ID || data["repositoryId"] != w.Reference.ID {
			return "", journalentity.ErrNotFound
		}
		name, ok := data["name"].(string)
		if !ok || name == "" {
			return "", journalentity.ErrNotFound
		}
		return name, nil
	}
	if w.ObjectID != "" {
		if !repositoryDomainEntity.ValidID(w.ObjectID) {
			return journalentity.ErrInput
		}
		object := repo.Collection("objects").Doc(w.ObjectID)
		w.ObjectName, err = readName(object)
		if err != nil {
			return err
		}
	}
	w.SpaceNames = nil
	seen := map[string]bool{}
	for _, id := range w.SpaceIDs {
		if !repositoryDomainEntity.ValidID(id) || seen[id] {
			return journalentity.ErrInput
		}
		seen[id] = true
		name, err := readName(repo.Collection("spaces").Doc(id))
		if err != nil {
			return err
		}
		w.SpaceNames = append(w.SpaceNames, name)
	}
	return nil
}

func (s *Store) Create(ctx context.Context, actorID, idempotencyKey string, d journalentity.Draft) (journalentity.Entry, error) {
	if d.GroupID != "" || !repositoryDomainEntity.ValidID(d.RepositoryID) {
		return journalentity.Entry{}, journalentity.ErrInput
	}
	owner, e := s.RepositoryOwner(ctx, actorID, d.RepositoryID)
	if e != nil {
		return journalentity.Entry{}, missing(e)
	}
	scope := d
	scope.GroupID = owner
	if e := s.checkRefs(ctx, actorID, scope, false, map[journalentity.Reference]error{}); e != nil {
		return journalentity.Entry{}, e
	}
	entryID := firestorekey.Digest([]byte(actorID + "\x00" + idempotencyKey))[:32]
	body, e := json.Marshal(draftData(d))
	if e != nil {
		return journalentity.Entry{}, e
	}
	claim := receipt{Repository: d.RepositoryID, Hash: firestorekey.Digest(body)}
	entry := journalentity.Entry{Draft: d, ID: entryID, CreatorID: actorID, CreatedAt: time.Now().UTC().Truncate(time.Microsecond)}
	order := d.WorkDate.Format(time.DateOnly) + "/" + entry.CreatedAt.Format("20060102T150405.000000Z") + "/" + entryID
	target := s.repositoryEntries(d.RepositoryID).Doc(entryID)
	locator := s.root().Collection("journalReceipts").Doc(entryID)
	e = s.Client.RunTransaction(ctx, func(_ context.Context, tx *db.Transaction) error {
		prev, e := tx.Get(locator)
		if e == nil {
			var saved receipt
			if e = prev.DataTo(&saved); e != nil {
				return e
			}
			if saved != claim {
				return journalentity.ErrConflict
			}
			snap, e := tx.Get(target)
			if e != nil {
				return e
			}
			entry, _, e = decode(snap)
			return e
		}
		if status.Code(e) != codes.NotFound {
			return e
		}
		snap, err := tx.Get(s.root().Collection("repositories").Doc(d.RepositoryID))
		if err != nil {
			return missing(err)
		}
		if snap.Data()["ownerId"] != scope.GroupID || snap.Data()["status"] != "ACTIVE" {
			return journalentity.ErrNotFound
		}
		// Copy rows so retries never mutate the caller's draft or its request hash.
		entry.Today = append([]journalentity.Work(nil), d.Today...)
		entry.Tomorrow = append([]journalentity.Work(nil), d.Tomorrow...)
		for _, rows := range [][]journalentity.Work{entry.Today, entry.Tomorrow} {
			for i := range rows {
				if e = s.snapshotWork(tx, scope.GroupID, &rows[i]); e != nil {
					return e
				}
			}
		}
		payload, e := json.Marshal(entryData(entry))
		if e != nil {
			return e
		}
		if e = tx.Create(locator, claim); e != nil {
			return e
		}
		return tx.Create(target, record{Payload: string(payload), Order: order})
	})
	return entry, e
}

func decode(snap *db.DocumentSnapshot) (journalentity.Entry, string, error) {
	var row record
	var item journalentity.Entry
	if e := snap.DataTo(&row); e != nil {
		return item, "", e
	}
	var payload entryPayload
	e := json.Unmarshal([]byte(row.Payload), &payload)
	return payload.model(), row.Order, e
}

func (s *Store) Get(ctx context.Context, actor, id string) (journalentity.Entry, error) {
	locator, e := s.root().Collection("journalReceipts").Doc(id).Get(ctx)
	if e != nil {
		return journalentity.Entry{}, missing(e)
	}
	var r receipt
	if e = locator.DataTo(&r); e != nil {
		return journalentity.Entry{}, e
	}
	var entries *db.CollectionRef
	owner := r.Group
	if r.Repository != "" {
		repositoryOwner, err := s.RepositoryOwner(ctx, actor, r.Repository)
		if err != nil {
			return journalentity.Entry{}, missing(err)
		}
		owner = repositoryOwner
		entries = s.repositoryEntries(r.Repository)
	} else {
		if e = s.group(ctx, actor, r.Group); e != nil {
			return journalentity.Entry{}, e
		}
		entries = s.entries(r.Group)
	}
	snap, e := entries.Doc(id).Get(ctx)
	if e != nil {
		return journalentity.Entry{}, missing(e)
	}
	item, _, e := decode(snap)
	if e == nil {
		scope := item.Draft
		scope.GroupID = owner
		e = s.checkRefs(ctx, actor, scope, true, map[journalentity.Reference]error{})
	}
	return item, e
}

func (s *Store) List(ctx context.Context, actor string, f journaldto.Filter) (journaldto.Page, error) {
	const size = 20
	scopeKey := actor + "\x00" + f.GroupID + "\x00" + f.WorkDate
	owner := ""
	if f.RepositoryID != "" {
		if f.GroupID != "" || !repositoryDomainEntity.ValidID(f.RepositoryID) {
			return journaldto.Page{}, journalentity.ErrInput
		}
		repositoryOwner, err := s.RepositoryOwner(ctx, actor, f.RepositoryID)
		if err != nil {
			return journaldto.Page{}, missing(err)
		}
		owner = repositoryOwner
		scopeKey += "\x00" + f.RepositoryID
	}
	scope := firestorekey.Digest([]byte(scopeKey)) + "/"
	after := ""
	if f.Cursor != "" {
		raw, e := base64.RawURLEncoding.DecodeString(f.Cursor)
		if e != nil || !strings.HasPrefix(string(raw), scope) {
			return journaldto.Page{}, journalentity.ErrInput
		}
		after = strings.TrimPrefix(string(raw), scope)
	}
	collections := []*db.CollectionRef{}
	if f.RepositoryID != "" {
		collections = append(collections, s.repositoryEntries(f.RepositoryID))
	} else if f.GroupID != "" {
		if e := s.group(ctx, actor, f.GroupID); e != nil {
			return journaldto.Page{}, e
		}
		collections = append(collections, s.entries(f.GroupID))
	} else {
		cursor := ""
		for {
			p, e := s.Groups.List(ctx, actor, cursor)
			if e != nil {
				return journaldto.Page{}, e
			}
			for _, g := range p.Items {
				collections = append(collections, s.entries(g.ID))
			}
			cursor = p.NextCursor
			if cursor == "" {
				break
			}
		}
	}
	type ordered struct {
		item  journalentity.Entry
		order string
	}
	rows := []ordered{}
	for _, entries := range collections {
		q := entries.OrderBy("order", db.Desc).Limit(size + 1)
		if f.WorkDate != "" {
			q = q.Where("order", ">=", f.WorkDate+"/").Where("order", "<", f.WorkDate+"0")
		}
		if after != "" {
			q = q.StartAfter(after)
		}
		snaps, e := q.Documents(ctx).GetAll()
		if e != nil {
			return journaldto.Page{}, e
		}
		for _, snap := range snaps {
			item, order, e := decode(snap)
			if e != nil {
				return journaldto.Page{}, e
			}
			rows = append(rows, ordered{item, order})
		}
		sort.Slice(rows, func(i, j int) bool { return rows[i].order > rows[j].order })
		if len(rows) > size+1 {
			rows = rows[:size+1]
		}
	}
	page := journaldto.Page{}
	if len(rows) > size {
		page.NextCursor = base64.RawURLEncoding.EncodeToString([]byte(scope + rows[size-1].order))
		rows = rows[:size]
	}
	caches := map[string]map[journalentity.Reference]error{}
	for _, r := range rows {
		scope := r.item.Draft
		if f.RepositoryID != "" {
			scope.GroupID = owner
		}
		g := scope.GroupID
		if caches[g] == nil {
			caches[g] = map[journalentity.Reference]error{}
		}
		if e := s.checkRefs(ctx, actor, scope, true, caches[g]); e != nil {
			return journaldto.Page{}, e
		}
		page.Items = append(page.Items, r.item)
	}
	return page, nil
}
