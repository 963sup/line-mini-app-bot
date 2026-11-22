package teamdb

import (
	"context"
	"encoding/base64"
	"strings"

	"cloud.google.com/go/firestore"
	"github.com/963sup/line-mini-app-bot/internal/business/internal/adapter/firestore"
	groupdto "github.com/963sup/line-mini-app-bot/internal/linegroup/application/dto"
	groupport "github.com/963sup/line-mini-app-bot/internal/linegroup/application/port"
	groupentity "github.com/963sup/line-mini-app-bot/internal/linegroup/domain/entity"
	groupid "github.com/963sup/line-mini-app-bot/internal/linegroup/domain/valueobject"
	firestorekey "github.com/963sup/line-mini-app-bot/internal/platform/firestore/key"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// GroupCatalog serves the existing group HTTP contract from formal Team data.
type GroupCatalog struct{ Store }

func (s GroupCatalog) Get(ctx context.Context, actor, id string) (groupentity.Group, error) {
	if actor == "" || !groupid.ValidID(id) {
		return groupentity.Group{}, groupentity.ErrNotFound
	}
	result, err := orgstore.Run(ctx, s.DB, s.Channel, func(tx *firestore.Transaction, root *firestore.DocumentRef) (*groupentity.Group, error) {
		return s.readGroup(tx, root, actor, id)
	})
	if err != nil {
		return groupentity.Group{}, err
	}
	if result == nil {
		return groupentity.Group{}, groupentity.ErrNotFound
	}
	return *result, nil
}

func (s GroupCatalog) readGroup(tx *firestore.Transaction, root *firestore.DocumentRef, actor, id string) (*groupentity.Group, error) {
	if s.TeamMember == nil {
		return nil, groupport.ErrUnavailable
	}
	snap, err := tx.Get(root.Collection("teams").Doc(id))
	if status.Code(err) == codes.NotFound {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	var team orgstore.Document
	if err := snap.DataTo(&team); err != nil {
		return nil, err
	}
	if team.Disabled || team.Status != "ACTIVE" || team.MembershipMode != "membership" {
		return nil, nil
	}
	ok, err := s.TeamMember(tx, root, id, team.Scope, team.Manager, actor)
	if err != nil || !ok {
		return nil, err
	}
	return &groupentity.Group{
		ID: id, GroupID: id, Name: team.Name, Status: team.Status,
		Version: snap.UpdateTime.UnixMicro(), CreatedAt: snap.CreateTime, UpdatedAt: snap.UpdateTime,
	}, nil
}

func (s GroupCatalog) List(ctx context.Context, actor, cursor string) (groupdto.Page, error) {
	if actor == "" {
		return groupdto.Page{}, groupentity.ErrNotFound
	}
	scope := firestorekey.JSONDigest([]string{s.Channel, actor})
	var after string
	if cursor != "" {
		raw, err := base64.RawURLEncoding.DecodeString(cursor)
		prefix, id, ok := strings.Cut(string(raw), ":")
		if len(cursor) > 1024 || err != nil || !ok || prefix != scope || id == "" || strings.Contains(id, "/") {
			return groupdto.Page{}, groupport.ErrCursor
		}
		after = id
	}
	result, err := orgstore.Run(ctx, s.DB, s.Channel, func(tx *firestore.Transaction, root *firestore.DocumentRef) (*groupdto.Page, error) {
		page := &groupdto.Page{Items: []groupentity.Group{}}
		members := root.Collection("memberships")
		q := members.Where("userId", "==", actor).OrderBy(firestore.DocumentID, firestore.Asc)
		if after != "" {
			q = q.StartAfter(members.Doc(after))
		}
		docs, err := tx.Documents(q.Limit(11)).GetAll()
		if err != nil {
			return nil, err
		}
		if len(docs) > 10 {
			page.NextCursor = base64.RawURLEncoding.EncodeToString([]byte(scope + ":" + docs[9].Ref.ID))
			docs = docs[:10]
		}
		for _, doc := range docs {
			var candidate struct {
				Kind string `firestore:"resourceType"`
				ID   string `firestore:"resourceId"`
			}
			if err := doc.DataTo(&candidate); err != nil {
				return nil, err
			}
			if candidate.Kind != "team" || !groupid.ValidID(candidate.ID) {
				continue
			}
			group, err := s.readGroup(tx, root, actor, candidate.ID)
			if err != nil {
				return nil, err
			}
			if group != nil {
				page.Items = append(page.Items, *group)
			}
		}
		return page, nil
	})
	if err != nil {
		return groupdto.Page{}, err
	}
	return *result, nil
}
