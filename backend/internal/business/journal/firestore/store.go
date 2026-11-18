// Package firestore persists immutable journal entries and their creation receipts.
package firestore

import (
	"context"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"errors"
	"sort"
	"strings"
	"time"

	db "cloud.google.com/go/firestore"
	groupdomain "github.com/963sup/line-mini-app-bot/internal/business/group"
	"github.com/963sup/line-mini-app-bot/internal/business/journal"
	"github.com/963sup/line-mini-app-bot/internal/business/project"
	"github.com/963sup/line-mini-app-bot/internal/business/repository"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type Store struct {
	Client       *db.Client
	Channel      string
	Groups       *groupdomain.Service
	Repositories *repository.Service
	Projects     *project.Service
}
type record struct {
	Payload string `firestore:"payload"`
	Order   string `firestore:"order"`
}
type receipt struct {
	Group string `firestore:"groupId"`
	Hash  string `firestore:"hash"`
}

func digest(data []byte) string        { h := sha256.Sum256(data); return hex.EncodeToString(h[:]) }
func (s *Store) root() *db.DocumentRef { return s.Client.Collection("channels").Doc(s.Channel) }
func (s *Store) entries(group string) *db.CollectionRef {
	return s.root().Collection("journalGroups").Doc(group).Collection("entries")
}

func missing(e error) error {
	if status.Code(e) == codes.NotFound || errors.Is(e, groupdomain.ErrNotFound) || errors.Is(e, repository.ErrNotFound) || errors.Is(e, project.ErrNotFound) {
		return journal.ErrNotFound
	}
	return e
}

func (s *Store) group(ctx context.Context, actor, group string) error {
	_, e := s.Groups.Get(ctx, actor, group)
	return missing(e)
}

func (s *Store) reference(ctx context.Context, actor, group string, r journal.Reference) error {
	switch r.Kind {
	case "repository":
		item, e := s.Repositories.Get(ctx, actor, r.ID)
		if e != nil {
			return missing(e)
		}
		if item.OwnerID == group {
			return nil
		}
	case "project":
		item, e := s.Projects.Get(ctx, actor, group, r.ID)
		if e != nil {
			return missing(e)
		}
		if item.OwnerID == group {
			return nil
		}
	}
	return journal.ErrNotFound
}

func (s *Store) checkRefs(ctx context.Context, actor string, d journal.Draft, redact bool, cache map[journal.Reference]error) error {
	for _, rows := range [][]journal.Work{d.Today, d.Tomorrow} {
		for i := range rows {
			r := rows[i].Reference
			if r == nil {
				continue
			}
			e, ok := cache[*r]
			if !ok {
				e = s.reference(ctx, actor, d.GroupID, *r)
				cache[*r] = e
			}
			if redact && errors.Is(e, journal.ErrNotFound) {
				rows[i].Reference = nil
				continue
			}
			if e != nil {
				return e
			}
		}
	}
	return nil
}

func (s *Store) Create(ctx context.Context, actor, key string, d journal.Draft) (journal.Entry, error) {
	if e := s.group(ctx, actor, d.GroupID); e != nil {
		return journal.Entry{}, e
	}
	if e := s.checkRefs(ctx, actor, d, false, map[journal.Reference]error{}); e != nil {
		return journal.Entry{}, e
	}
	id := digest([]byte(actor + "\x00" + key))[:32]
	body, e := json.Marshal(d)
	if e != nil {
		return journal.Entry{}, e
	}
	claim := receipt{Group: d.GroupID, Hash: digest(body)}
	entry := journal.Entry{Draft: d, ID: id, CreatorID: actor, CreatedAt: time.Now().UTC().Truncate(time.Microsecond)}
	payload, e := json.Marshal(entry)
	if e != nil {
		return journal.Entry{}, e
	}
	order := d.WorkDate.Format(time.DateOnly) + "/" + entry.CreatedAt.Format("20060102T150405.000000Z") + "/" + id
	target := s.entries(d.GroupID).Doc(id)
	locator := s.root().Collection("journalReceipts").Doc(id)
	e = s.Client.RunTransaction(ctx, func(_ context.Context, tx *db.Transaction) error {
		prev, e := tx.Get(locator)
		if e == nil {
			var saved receipt
			if e = prev.DataTo(&saved); e != nil {
				return e
			}
			if saved != claim {
				return journal.ErrConflict
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
		if e = tx.Create(locator, claim); e != nil {
			return e
		}
		return tx.Create(target, record{Payload: string(payload), Order: order})
	})
	return entry, e
}

func decode(snap *db.DocumentSnapshot) (journal.Entry, string, error) {
	var row record
	var item journal.Entry
	if e := snap.DataTo(&row); e != nil {
		return item, "", e
	}
	e := json.Unmarshal([]byte(row.Payload), &item)
	return item, row.Order, e
}

func (s *Store) Get(ctx context.Context, actor, id string) (journal.Entry, error) {
	locator, e := s.root().Collection("journalReceipts").Doc(id).Get(ctx)
	if e != nil {
		return journal.Entry{}, missing(e)
	}
	var r receipt
	if e = locator.DataTo(&r); e != nil {
		return journal.Entry{}, e
	}
	if e = s.group(ctx, actor, r.Group); e != nil {
		return journal.Entry{}, e
	}
	snap, e := s.entries(r.Group).Doc(id).Get(ctx)
	if e != nil {
		return journal.Entry{}, missing(e)
	}
	item, _, e := decode(snap)
	if e == nil {
		e = s.checkRefs(ctx, actor, item.Draft, true, map[journal.Reference]error{})
	}
	return item, e
}

func (s *Store) List(ctx context.Context, actor string, f journal.Filter) (journal.Page, error) {
	const size = 20
	scope := digest([]byte(actor+"\x00"+f.GroupID+"\x00"+f.WorkDate)) + "/"
	after := ""
	if f.Cursor != "" {
		raw, e := base64.RawURLEncoding.DecodeString(f.Cursor)
		if e != nil || !strings.HasPrefix(string(raw), scope) {
			return journal.Page{}, journal.ErrInput
		}
		after = strings.TrimPrefix(string(raw), scope)
	}
	groups := []string{}
	if f.GroupID != "" {
		if e := s.group(ctx, actor, f.GroupID); e != nil {
			return journal.Page{}, e
		}
		groups = append(groups, f.GroupID)
	} else {
		cursor := ""
		for {
			p, e := s.Groups.List(ctx, actor, cursor)
			if e != nil {
				return journal.Page{}, e
			}
			for _, g := range p.Items {
				groups = append(groups, g.ID)
			}
			cursor = p.NextCursor
			if cursor == "" {
				break
			}
		}
	}
	type ordered struct {
		item  journal.Entry
		order string
	}
	rows := []ordered{}
	for _, g := range groups {
		q := s.entries(g).OrderBy("order", db.Desc).Limit(size + 1)
		if f.WorkDate != "" {
			q = q.Where("order", ">=", f.WorkDate+"/").Where("order", "<", f.WorkDate+"0")
		}
		if after != "" {
			q = q.StartAfter(after)
		}
		snaps, e := q.Documents(ctx).GetAll()
		if e != nil {
			return journal.Page{}, e
		}
		for _, snap := range snaps {
			item, order, e := decode(snap)
			if e != nil {
				return journal.Page{}, e
			}
			rows = append(rows, ordered{item, order})
		}
		sort.Slice(rows, func(i, j int) bool { return rows[i].order > rows[j].order })
		if len(rows) > size+1 {
			rows = rows[:size+1]
		}
	}
	page := journal.Page{}
	if len(rows) > size {
		page.NextCursor = base64.RawURLEncoding.EncodeToString([]byte(scope + rows[size-1].order))
		rows = rows[:size]
	}
	caches := map[string]map[journal.Reference]error{}
	for _, r := range rows {
		g := r.item.GroupID
		if caches[g] == nil {
			caches[g] = map[journal.Reference]error{}
		}
		if e := s.checkRefs(ctx, actor, r.item.Draft, true, caches[g]); e != nil {
			return journal.Page{}, e
		}
		page.Items = append(page.Items, r.item)
	}
	return page, nil
}
