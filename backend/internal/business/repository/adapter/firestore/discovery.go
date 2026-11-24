package firestore

import (
	"context"

	db "cloud.google.com/go/firestore"
	"github.com/963sup/line-mini-app-bot/internal/business/repository/application/dto"
	"github.com/963sup/line-mini-app-bot/internal/business/repository/domain/entity"
)

// DiscoverableCandidates uses an opt-in filter so legacy documents remain hidden.
func (s *Store) DiscoverableCandidates(ctx context.Context, after string) (dto.Page, error) {
	if !s.ready() {
		return dto.Page{}, entity.ErrUnavailable
	}
	q := s.channel().Collection("repositories").Where("discoverable", "==", true).OrderBy(db.DocumentID, db.Asc)
	if after != "" {
		if !entity.ValidID(after) {
			return dto.Page{}, entity.ErrCursor
		}
		q = q.StartAfter(after)
	}
	const size = 10
	docs, err := q.Limit(size + 1).Documents(ctx).GetAll()
	if err != nil {
		return dto.Page{}, err
	}
	page := dto.Page{}
	if len(docs) > size {
		page.NextCursor = docs[size-1].Ref.ID
		docs = docs[:size]
	}
	for _, doc := range docs {
		r, err := decode(doc, nil)
		if err != nil {
			return dto.Page{}, err
		}
		page.Items = append(page.Items, r)
	}
	return page, nil
}
