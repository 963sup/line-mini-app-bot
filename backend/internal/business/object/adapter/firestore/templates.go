package firestore

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"reflect"
	"time"

	db "cloud.google.com/go/firestore"
	"github.com/963sup/line-mini-app-bot/internal/business/object/application"
	"github.com/963sup/line-mini-app-bot/internal/business/object/application/dto"
	"github.com/963sup/line-mini-app-bot/internal/business/object/application/port"
	"github.com/963sup/line-mini-app-bot/internal/business/object/domain/entity"
	"github.com/963sup/line-mini-app-bot/internal/business/object/domain/template"
	repository "github.com/963sup/line-mini-app-bot/internal/business/repository/domain/entity"
	"github.com/963sup/line-mini-app-bot/internal/platform/firestorekey"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

var _ port.TemplateStore = (*Store)(nil)

type templateNodeDocument struct {
	ID          string `firestore:"id"`
	ParentID    string `firestore:"parentId"`
	Name        string `firestore:"name"`
	Description string `firestore:"description"`
	Type        string `firestore:"type"`
	Quantity    string `firestore:"quantity"`
	Unit        string `firestore:"unit"`
}
type templateDocument struct {
	SchemaVersion int64                  `firestore:"schemaVersion"`
	ID            string                 `firestore:"id"`
	RepositoryID  string                 `firestore:"repositoryId"`
	Name          string                 `firestore:"name"`
	CreatorID     string                 `firestore:"creatorId"`
	Version       int64                  `firestore:"version"`
	Nodes         []templateNodeDocument `firestore:"nodes"`
	CreatedAt     time.Time              `firestore:"createdAt"`
}

func templateEncode(t dto.Template) templateDocument {
	nodes := make([]templateNodeDocument, len(t.Nodes))
	for i, n := range t.Nodes {
		nodes[i] = templateNodeDocument{n.ID, n.ParentID, n.Name, n.Description, n.Type, n.Quantity, n.Unit}
	}
	return templateDocument{1, t.ID, t.RepositoryID, t.Name, t.CreatorID, t.Version, nodes, t.CreatedAt}
}

func templateDecode(snap *db.DocumentSnapshot, err error, rid string) (dto.Template, error) {
	if status.Code(err) == codes.NotFound {
		return dto.Template{}, entity.ErrNotFound
	}
	if err != nil {
		return dto.Template{}, err
	}
	var d templateDocument
	if e := snap.DataTo(&d); e != nil {
		return dto.Template{}, e
	}
	nodes := make([]template.Node, len(d.Nodes))
	for i, n := range d.Nodes {
		nodes[i] = template.Node{ID: n.ID, ParentID: n.ParentID, Name: n.Name, Description: n.Description, Type: n.Type, Quantity: n.Quantity, Unit: n.Unit}
	}
	raw := dto.TemplateCreation{Name: d.Name, Nodes: nodes}
	normalized, e := application.NormalizeTemplate(raw)
	if e != nil || !reflect.DeepEqual(raw, normalized) || d.SchemaVersion != 1 || d.Version != 1 || !repository.ValidID(d.ID) || d.ID != snap.Ref.ID || d.RepositoryID != rid || d.CreatorID == "" || d.CreatedAt.IsZero() {
		return dto.Template{}, entity.ErrUnavailable
	}
	return dto.Template{ID: d.ID, RepositoryID: rid, Name: d.Name, CreatorID: d.CreatorID, Version: d.Version, Nodes: nodes, CreatedAt: d.CreatedAt}, nil
}

func (s *Store) CreateTemplate(ctx context.Context, r repository.Repository, actor, key string, in dto.TemplateCreation) (result dto.Template, err error) {
	defer func() { err = translate(err) }()
	if !s.ready() {
		return result, entity.ErrUnavailable
	}
	parent := s.repositoryRef(r.ID)
	ref := parent.Collection("objectTemplates").NewDoc()
	// Match existing Object identity format, rather than Firestore's auto-ID format.
	ref = parent.Collection("objectTemplates").Doc(firestorekey.JSONDigest([]string{ref.ID, actor, key})[:32])
	rec := parent.Collection("objectTemplateCreationReceipts").Doc(firestorekey.JSONDigest([]string{actor, key}))
	payload := firestorekey.JSONDigest(in)
	now := time.Now().UTC().Truncate(time.Microsecond)
	err = s.Client.RunTransaction(ctx, func(_ context.Context, tx *db.Transaction) error {
		result = dto.Template{}
		if e := container(tx, parent, r); e != nil {
			return e
		}
		old, e := tx.Get(rec)
		if e == nil {
			var receipt receipt
			if e = old.DataTo(&receipt); e != nil {
				return e
			}
			if receipt.SchemaVersion != 1 || !repository.ValidID(receipt.ID) || len(receipt.PayloadHash) != 64 {
				return entity.ErrUnavailable
			}
			if receipt.PayloadHash != payload {
				return entity.ErrCreationConflict
			}
			snap, e := tx.Get(parent.Collection("objectTemplates").Doc(receipt.ID))
			result, e = templateDecode(snap, e, r.ID)
			if e == nil && result.CreatorID != actor {
				return entity.ErrUnavailable
			}
			return e
		}
		if status.Code(e) != codes.NotFound {
			return e
		}
		if _, e = template.Publish(r.ID, ref.ID, 1, in.Nodes); e != nil {
			return entity.ErrInput
		}
		result = dto.Template{ID: ref.ID, RepositoryID: r.ID, Name: in.Name, CreatorID: actor, Version: 1, Nodes: in.Nodes, CreatedAt: now}
		if e = tx.Create(ref, templateEncode(result)); e != nil {
			return e
		}
		return tx.Create(rec, receipt{1, ref.ID, payload})
	})
	if err != nil {
		return dto.Template{}, err
	}
	return result, nil
}

func (s *Store) GetTemplate(ctx context.Context, r repository.Repository, id string) (result dto.Template, err error) {
	defer func() { err = translate(err) }()
	if !s.ready() {
		return result, entity.ErrUnavailable
	}
	snap, e := s.repositoryRef(r.ID).Collection("objectTemplates").Doc(id).Get(ctx)
	return templateDecode(snap, e, r.ID)
}

func (s *Store) ListTemplates(ctx context.Context, r repository.Repository, actor, token string) (result dto.TemplatePage, err error) {
	defer func() { err = translate(err) }()
	result.Items = []dto.Template{}
	if !s.ready() {
		return result, entity.ErrUnavailable
	}
	scope := firestorekey.JSONDigest([]string{s.ChannelID, r.ID, actor, "objectTemplates"})
	c := cursor{}
	if token != "" {
		b, e := base64.RawURLEncoding.DecodeString(token)
		if e != nil || json.Unmarshal(b, &c) != nil || c.Scope != scope || !repository.ValidID(c.ID) {
			return result, entity.ErrInput
		}
	}
	q := s.repositoryRef(r.ID).Collection("objectTemplates").OrderBy(db.DocumentID, db.Asc)
	if c.ID != "" {
		q = q.StartAfter(c.ID)
	}
	docs, e := q.Limit(21).Documents(ctx).GetAll()
	if e != nil {
		return result, e
	}
	more := len(docs) > 20
	if more {
		docs = docs[:20]
	}
	for _, snap := range docs {
		t, e := templateDecode(snap, nil, r.ID)
		if e != nil {
			return dto.TemplatePage{}, e
		}
		result.Items = append(result.Items, t)
		c = cursor{scope, t.ID}
	}
	if more {
		b, _ := json.Marshal(c)
		result.NextCursor = base64.RawURLEncoding.EncodeToString(b)
	}
	return result, nil
}

type templateInstanceNodeDocument struct {
	ObjectID string `firestore:"objectId"`
	ParentID string `firestore:"parentId"`
	Name     string `firestore:"name"`
}

type templateInstanceReceipt struct {
	Nodes           []templateInstanceNodeDocument `firestore:"nodes"`
	SchemaVersion   int64                          `firestore:"schemaVersion"`
	PayloadHash     string                         `firestore:"payloadHash"`
	RootObjectID    string                         `firestore:"rootObjectId"`
	ObjectIDs       []string                       `firestore:"objectIds"`
	TemplateID      string                         `firestore:"templateId"`
	TemplateVersion int64                          `firestore:"templateVersion"`
}

func (s *Store) InstantiateTemplate(ctx context.Context, r repository.Repository, actor, key string, snapshot dto.Template, parentID string, plan []entity.Content) (result dto.TemplateInstance, err error) {
	defer func() { err = translate(err) }()
	if !s.ready() {
		return result, entity.ErrUnavailable
	}
	if len(plan) != len(snapshot.Nodes) || len(plan) == 0 || len(plan) > application.MaxTemplateNodes || snapshot.RepositoryID != r.ID || snapshot.Version != 1 {
		return result, entity.ErrInput
	}
	parent := s.repositoryRef(r.ID)
	rec := parent.Collection("objectTemplateInstanceReceipts").Doc(firestorekey.JSONDigest([]string{actor, key}))
	payload := firestorekey.JSONDigest(struct {
		TemplateID string
		Version    int64
		ParentID   string
	}{snapshot.ID, snapshot.Version, parentID})
	ids := make(map[string]string, len(plan))
	refs := make([]*db.DocumentRef, len(plan))
	rootID := ""
	for i, n := range snapshot.Nodes {
		seed := parent.Collection("objects").NewDoc().ID
		id := firestorekey.JSONDigest([]string{seed, actor, key, n.ID})[:32]
		refs[i] = parent.Collection("objects").Doc(id)
		ids[n.ID] = id
		if n.ParentID == "" {
			rootID = id
		}
	}
	now := time.Now().UTC().Truncate(time.Microsecond)
	err = s.Client.RunTransaction(ctx, func(_ context.Context, tx *db.Transaction) error {
		result = dto.TemplateInstance{}
		if e := container(tx, parent, r); e != nil {
			return e
		}
		old, e := tx.Get(rec)
		if e == nil {
			var receipt templateInstanceReceipt
			if e = old.DataTo(&receipt); e != nil {
				return e
			}
			if receipt.SchemaVersion != 1 || !repository.ValidID(receipt.RootObjectID) || len(receipt.PayloadHash) != 64 || len(receipt.ObjectIDs) == 0 || receipt.TemplateVersion != 1 || !repository.ValidID(receipt.TemplateID) {
				return entity.ErrUnavailable
			}
			if receipt.PayloadHash != payload {
				return entity.ErrCreationConflict
			}
			result = dto.TemplateInstance{RootObjectID: receipt.RootObjectID, ObjectIDs: receipt.ObjectIDs, TemplateID: receipt.TemplateID, TemplateVersion: receipt.TemplateVersion}
			for _, n := range receipt.Nodes {
				result.Nodes = append(result.Nodes, dto.TemplateInstanceNode{ObjectID: n.ObjectID, ParentID: n.ParentID, Name: n.Name})
			}
			if len(result.Nodes) != len(result.ObjectIDs) {
				return entity.ErrUnavailable
			}
			return nil
		}
		if status.Code(e) != codes.NotFound {
			return e
		}
		saved, e := tx.Get(parent.Collection("objectTemplates").Doc(snapshot.ID))
		current, e := templateDecode(saved, e, r.ID)
		if e != nil {
			return e
		}
		if !reflect.DeepEqual(current, snapshot) {
			return entity.ErrConflict
		}
		if parentID != "" {
			snap, e := tx.Get(parent.Collection("objects").Doc(parentID))
			if _, e = decode(snap, e, r.ID); e != nil {
				return e
			}
		}
		// All reads precede every write. Objects and the replay receipt commit together.
		result = dto.TemplateInstance{RootObjectID: rootID, TemplateID: snapshot.ID, TemplateVersion: snapshot.Version, ObjectIDs: make([]string, len(plan))}
		for i, n := range snapshot.Nodes {
			c := plan[i]
			c.ParentID = parentID
			if n.ParentID != "" {
				c.ParentID = ids[n.ParentID]
			}
			c, e = entity.Normalize(c)
			if e != nil {
				return e
			}
			object := entity.Object{Content: c, ID: refs[i].ID, RepositoryID: r.ID, CreatorID: actor, Status: initialStatus(c), Version: 1, CreatedAt: now, UpdatedAt: now, TemplateID: snapshot.ID, TemplateVersion: snapshot.Version, TemplateNodeID: n.ID}
			if e = tx.Create(refs[i], encode(object)); e != nil {
				return e
			}
			result.ObjectIDs[i] = object.ID
			result.Nodes = append(result.Nodes, dto.TemplateInstanceNode{ObjectID: object.ID, ParentID: c.ParentID, Name: c.Name})
		}
		nodes := make([]templateInstanceNodeDocument, len(result.Nodes))
		for i, n := range result.Nodes {
			nodes[i] = templateInstanceNodeDocument{n.ObjectID, n.ParentID, n.Name}
		}
		return tx.Create(rec, templateInstanceReceipt{Nodes: nodes, SchemaVersion: 1, PayloadHash: payload, RootObjectID: rootID, ObjectIDs: result.ObjectIDs, TemplateID: snapshot.ID, TemplateVersion: snapshot.Version})
	})
	if err != nil {
		return dto.TemplateInstance{}, err
	}
	return result, nil
}
