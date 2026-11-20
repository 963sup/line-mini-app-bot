package firestore

import (
	repositoryDomainEntity "github.com/963sup/line-mini-app-bot/internal/business/repository/domain/entity"
	repositoryDomainEntityWiki "github.com/963sup/line-mini-app-bot/internal/business/repository/domain/entity/wiki"
)

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"time"

	db "cloud.google.com/go/firestore"

	"github.com/963sup/line-mini-app-bot/internal/persistence/firestorekey"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func readWiki(snap *db.DocumentSnapshot, err error) (p repositoryDomainEntityWiki.Page, e error) {
	if status.Code(err) == codes.NotFound {
		return p, repositoryDomainEntity.ErrNotFound
	}
	if err != nil {
		return p, err
	}
	e = snap.DataTo(&p)
	return
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

func (s *Store) SaveWiki(ctx context.Context, r repositoryDomainEntity.Repository, actor, id, key string, expected int64, title, body string) (result repositoryDomainEntityWiki.Page, err error) {
	title, err = repositoryDomainEntityWiki.Validate(title, body)
	if err != nil {
		return
	}
	create := id == ""
	if create {
		id = firestorekey.Encode("wiki", actor, key)[:32]
	}
	root := s.repositoryRef(r.ID)
	doc := root.Collection("wikiPages").Doc(id)
	names := root.Collection("wikiTitles")
	name := names.Doc(firestorekey.Encode(strings.ToLower(title)))
	now := time.Now().UTC()
	err = s.Client.RunTransaction(ctx, func(_ context.Context, tx *db.Transaction) error {
		scope, e := decode(tx.Get(root))
		if e != nil {
			return e
		}
		if scope.OwnerID != r.OwnerID || scope.Status != "ACTIVE" {
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
		if create && exists {
			first, e := readWiki(tx.Get(doc.Collection("revisions").Doc("0000000000000001")))
			if e != nil {
				return e
			}
			if err := first.CheckCreationReplay(actor, title, body); err != nil {
				return err
			}
			result = current
			return nil
		}
		if !create {
			if !exists {
				return repositoryDomainEntity.ErrNotFound
			}
			changed, err := current.NeedsRevision(expected, title, body)
			if err != nil {
				return err
			}
			if !changed {
				result = current
				return nil
			}
		}
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
		result = repositoryDomainEntityWiki.Page{ID: id, Title: title, Body: body, AuthorID: actor, Version: current.Version + 1, UpdatedAt: now}
		if e = tx.Set(doc, result); e != nil {
			return e
		}
		return tx.Create(doc.Collection("revisions").Doc(fmt.Sprintf("%016d", result.Version)), result)
	})
	return
}
