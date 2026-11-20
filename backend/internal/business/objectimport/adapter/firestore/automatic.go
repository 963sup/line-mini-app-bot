package firestore

import (
	repositoryDomainEntity "github.com/963sup/line-mini-app-bot/internal/business/repository/domain/entity"
)

import (
	objectimportApplicationPortAutomatic "github.com/963sup/line-mini-app-bot/internal/business/objectimport/application/port/automatic"
	objectimportDomainEntity "github.com/963sup/line-mini-app-bot/internal/business/objectimport/domain/entity"
	objectimportDomainEntityAutomatic "github.com/963sup/line-mini-app-bot/internal/business/objectimport/domain/entity/automatic"
)

import (
	"context"
	"encoding/json"
	"errors"
	"slices"
	"strings"
	"time"

	db "cloud.google.com/go/firestore"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// Private version-one adapter DTOs keep storage independent of domain Go names.
type storedAutomaticRow struct {
	ID             string `json:"id"`
	Name           string `json:"name"`
	Description    string `json:"description"`
	Type           string `json:"type"`
	Quantity       string `json:"quantity"`
	Unit           string `json:"unit"`
	MaterialCode   string `json:"materialCode"`
	SourceItemNo   string `json:"sourceItemNo"`
	SourcePages    string `json:"sourcePages"`
	ResultObjectID string `json:"resultObjectId"`
}
type storedAutomaticRun struct {
	ID             string               `json:"id"`
	RepositoryID   string               `json:"repositoryId"`
	CreatorID      string               `json:"creatorId"`
	IntentHash     string               `json:"intentHash"`
	SourceDocument string               `json:"sourceDocument"`
	SourceRevision string               `json:"sourceRevision"`
	SourceHash     string               `json:"sourceHash"`
	Status         string               `json:"status"`
	Reason         string               `json:"reason"`
	Rows           []storedAutomaticRow `json:"rows"`
	CreatedAt      time.Time            `json:"createdAt"`
	UpdatedAt      time.Time            `json:"updatedAt"`
}

func toStoredRun(j objectimportDomainEntityAutomatic.Run) storedAutomaticRun {
	d := storedAutomaticRun{ID: j.ID, RepositoryID: j.RepositoryID, CreatorID: j.CreatorID, IntentHash: j.IntentHash, SourceDocument: j.SourceDocument, SourceRevision: j.SourceRevision, SourceHash: j.SourceHash, Status: j.Status, Reason: j.Reason, Rows: []storedAutomaticRow{}, CreatedAt: j.CreatedAt, UpdatedAt: j.UpdatedAt}
	for _, r := range j.Rows {
		d.Rows = append(d.Rows, storedAutomaticRow{ID: r.ID, Name: r.Name, Description: r.Description, Type: r.Type, Quantity: r.Quantity, Unit: r.Unit, MaterialCode: r.MaterialCode, SourceItemNo: r.SourceItemNo, SourcePages: r.SourcePages, ResultObjectID: r.ResultObjectID})
	}
	return d
}

func fromStoredRun(d storedAutomaticRun) objectimportDomainEntityAutomatic.Run {
	j := objectimportDomainEntityAutomatic.Run{ID: d.ID, RepositoryID: d.RepositoryID, CreatorID: d.CreatorID, IntentHash: d.IntentHash, SourceDocument: d.SourceDocument, SourceRevision: d.SourceRevision, SourceHash: d.SourceHash, Status: d.Status, Reason: d.Reason, Rows: []objectimportDomainEntityAutomatic.Row{}, CreatedAt: d.CreatedAt, UpdatedAt: d.UpdatedAt}
	for _, r := range d.Rows {
		j.Rows = append(j.Rows, objectimportDomainEntityAutomatic.Row{ID: r.ID, Fields: objectimportDomainEntity.Fields{Name: r.Name, Description: r.Description, Type: r.Type, Quantity: r.Quantity, Unit: r.Unit, MaterialCode: r.MaterialCode}, SourceItemNo: r.SourceItemNo, SourcePages: r.SourcePages, ResultObjectID: r.ResultObjectID})
	}
	return j
}

func encodeAutomatic(j objectimportDomainEntityAutomatic.Run) (document, error) {
	if objectimportDomainEntityAutomatic.ValidateStored(j) != nil {
		return document{}, objectimportDomainEntity.ErrUnavailable
	}
	d := toStoredRun(j)
	projected := d
	projected.Rows = slices.Clone(d.Rows)
	projected.Reason = "creation_unavailable"
	projected.Status = "processing"
	projected.UpdatedAt = time.Date(9999, 12, 31, 23, 59, 59, 999999999, time.UTC)
	for i := range projected.Rows {
		projected.Rows[i].ResultObjectID = strings.Repeat("f", 32)
	}
	b, e := json.Marshal(projected)
	if e != nil || len(b)+4096 > objectimportDomainEntity.MaxDraftBytes {
		return document{}, objectimportDomainEntity.ErrInput
	}
	b, e = json.Marshal(d)
	if e != nil {
		return document{}, objectimportDomainEntity.ErrUnavailable
	}
	return document{SchemaVersion: 1, Payload: string(b)}, nil
}

// AutomaticStore uses the existing client/container boundary without changing manual drafts.
type AutomaticStore struct{ Store }

var _ objectimportApplicationPortAutomatic.Store = (*AutomaticStore)(nil)

func (s *AutomaticStore) Update(ctx context.Context, r repositoryDomainEntity.Repository, actor, id string, apply func(*objectimportDomainEntityAutomatic.Run) (objectimportDomainEntityAutomatic.Run, error)) (result objectimportDomainEntityAutomatic.Run, err error) {
	defer func() {
		if !errors.Is(err, objectimportDomainEntityAutomatic.ErrTypeRequired) {
			err = translate(err)
		}
	}()
	if !s.ready() {
		return result, objectimportDomainEntity.ErrUnavailable
	}
	if actor == "" || !repositoryDomainEntity.ValidID(id) {
		return result, objectimportDomainEntity.ErrNotFound
	}
	parent := s.repositoryRef(r.ID)
	ref := parent.Collection("objectImportRuns").Doc(id)
	err = s.Client.RunTransaction(ctx, func(_ context.Context, tx *db.Transaction) error {
		result = objectimportDomainEntityAutomatic.Run{}
		if e := checkActiveRepository(tx, parent, r); e != nil {
			return e
		}
		snap, e := tx.Get(ref)
		var old *objectimportDomainEntityAutomatic.Run
		if e == nil {
			var d document
			if snap.DataTo(&d) != nil || d.SchemaVersion != 1 || len(d.Payload) > objectimportDomainEntity.MaxDraftBytes {
				return objectimportDomainEntity.ErrUnavailable
			}
			var stored storedAutomaticRun
			if json.Unmarshal([]byte(d.Payload), &stored) != nil {
				return objectimportDomainEntity.ErrUnavailable
			}
			j := fromStoredRun(stored)
			if j.ID != id || j.RepositoryID != r.ID || objectimportDomainEntityAutomatic.ValidateStored(j) != nil {
				return objectimportDomainEntity.ErrUnavailable
			}
			if j.CreatorID != actor {
				return objectimportDomainEntity.ErrNotFound
			}
			old = &j
		} else if status.Code(e) != codes.NotFound {
			return e
		}
		result, e = apply(old)
		if e != nil {
			return e
		}
		if result.ID != id || result.RepositoryID != r.ID || result.CreatorID != actor {
			return objectimportDomainEntity.ErrInput
		}
		d, e := encodeAutomatic(result)
		if e != nil {
			return e
		}
		if old != nil {
			prior, e := encodeAutomatic(*old)
			if e != nil {
				return e
			}
			if prior.Payload == d.Payload {
				return nil
			}
		}
		return tx.Set(ref, d)
	})
	if err != nil {
		return objectimportDomainEntityAutomatic.Run{}, err
	}
	return result, nil
}
