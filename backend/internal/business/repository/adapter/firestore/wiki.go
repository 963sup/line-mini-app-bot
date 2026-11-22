package firestore

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"time"

	db "cloud.google.com/go/firestore"

	wikiport "github.com/963sup/line-mini-app-bot/internal/business/repository/application/port/wiki"
	repositoryDomainEntity "github.com/963sup/line-mini-app-bot/internal/business/repository/domain/entity"
	repositoryDomainEntityWiki "github.com/963sup/line-mini-app-bot/internal/business/repository/domain/entity/wiki"
	"github.com/963sup/line-mini-app-bot/internal/platform/firestore/key"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// wikiDocument retains the stored schema for current pages and revisions.
type wikiDocument struct {
	ID        string    `firestore:"id"`
	Title     string    `firestore:"title"`
	Body      string    `firestore:"body"`
	AuthorID  string    `firestore:"authorId"`
	Version   int64     `firestore:"version"`
	UpdatedAt time.Time `firestore:"updatedAt"`
}

func readWiki(snap *db.DocumentSnapshot, err error) (p repositoryDomainEntityWiki.Page, e error) {
	if status.Code(err) == codes.NotFound {
		return p, repositoryDomainEntity.ErrNotFound
	}
	if err != nil {
		return p, err
	}
	var stored wikiDocument
	if e = snap.DataTo(&stored); e != nil {
		return p, e
	}
	return repositoryDomainEntityWiki.Page(stored), nil
}

func (s *Store) GetWiki(ctx context.Context, r repositoryDomainEntity.Repository, id string) (repositoryDomainEntityWiki.Page, error) {
	return readWiki(s.repositoryRef(r.ID).Collection("wikiPages").Doc(id).Get(ctx))
}

func (s *Store) ListWiki(ctx context.Context, r repositoryDomainEntity.Repository, actor, id, cursor string) ([]repositoryDomainEntityWiki.Page, string, error) {
	pages := s.repositoryRef(r.ID).Collection("wikiPages")
	q := pages.OrderBy(db.DocumentID, db.Asc).Select("id", "title", "authorId", "version", "updatedAt")
	if id != "" {
		doc := pages.Doc(id)
		if _, e := readWiki(doc.Get(ctx)); e != nil {
			return nil, "", e
		}
		q = doc.Collection("revisions").OrderBy("version", db.Desc)
	}
	scope := []string{s.ChannelID, r.ID, actor, id}
	if cursor != "" {
		last, ok := firestorekey.DecodeCursor(cursor, scope)
		if !ok {
			return nil, "", repositoryDomainEntityWiki.ErrInput
		}
		if id == "" {
			if !repositoryDomainEntity.ValidID(last) {
				return nil, "", repositoryDomainEntityWiki.ErrInput
			}
			q = q.StartAfter(last)
		} else if v, e := strconv.ParseInt(last, 10, 64); e != nil || v < 1 || fmt.Sprintf("%016d", v) != last {
			return nil, "", repositoryDomainEntityWiki.ErrInput
		} else {
			q = q.StartAfter(v)
		}
	}
	docs, e := q.Limit(21).Documents(ctx).GetAll()
	if e != nil {
		return nil, "", e
	}
	next := ""
	if len(docs) > 20 {
		docs = docs[:20]
		next = firestorekey.EncodeCursor(scope, docs[19].Ref.ID)
	}
	items := make([]repositoryDomainEntityWiki.Page, 0, len(docs))
	for _, d := range docs {
		p, e := readWiki(d, nil)
		if e != nil {
			return nil, "", e
		}
		items = append(items, p)
	}
	return items, next, nil
}

func (s *Store) WriteWiki(ctx context.Context, r repositoryDomainEntity.Repository, target wikiport.WriteTarget, decide wikiport.DecideWrite) (repositoryDomainEntityWiki.Page, error) {
	id := target.ID
	create := id == ""
	if create {
		id = firestorekey.Encode("wiki", target.Actor, target.CreationKey)[:32]
	}
	var result repositoryDomainEntityWiki.Page
	root := s.repositoryRef(r.ID)
	doc := root.Collection("wikiPages").Doc(id)
	names := root.Collection("wikiTitles")
	err := s.Client.RunTransaction(ctx, func(_ context.Context, tx *db.Transaction) error {
		scope, e := decode(tx.Get(root))
		if e != nil {
			return e
		}
		if !r.MatchesActive(r.ID, scope.OwnerID, scope.Status) {
			return repositoryDomainEntity.ErrNotFound
		}
		snap, e := tx.Get(doc)
		exists := e == nil
		if !exists && status.Code(e) != codes.NotFound {
			return e
		}
		var current repositoryDomainEntityWiki.Page
		if exists {
			current, e = readWiki(snap, nil)
			if e != nil {
				return e
			}
		}
		state := wikiport.WriteState{ID: id}
		if exists {
			state.Current = &current
			if create {
				first, err := readWiki(tx.Get(doc.Collection("revisions").Doc("0000000000000001")))
				if err != nil {
					return err
				}
				state.FirstRevision = &first
			}
		}
		next, changed, err := decide(state)
		if err != nil {
			return err
		}
		result = next
		if !changed {
			return nil
		}
		name := names.Doc(firestorekey.Encode(strings.ToLower(next.Title)))
		oldName := names.Doc(firestorekey.Encode(strings.ToLower(current.Title)))
		if !exists || oldName.ID != name.ID {
			_, e = tx.Get(name)
			if e == nil {
				return repositoryDomainEntityWiki.ErrConflict
			}
			if status.Code(e) != codes.NotFound {
				return e
			}
			if e = tx.Create(name, map[string]string{"pageId": id}); e != nil {
				return e
			}
			if exists {
				if e = tx.Delete(oldName); e != nil {
					return e
				}
			}
		}
		stored := wikiDocument(next)
		if e = tx.Set(doc, stored); e != nil {
			return e
		}
		return tx.Create(doc.Collection("revisions").Doc(fmt.Sprintf("%016d", next.Version)), stored)
	})
	if err != nil {
		return repositoryDomainEntityWiki.Page{}, err
	}
	return result, nil
}
