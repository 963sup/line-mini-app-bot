package firestore

import (
	"context"
	"errors"
	"time"

	"github.com/963sup/line-mini-app-bot/internal/persistence/grpcstatus"

	db "cloud.google.com/go/firestore"
	feature "github.com/963sup/line-mini-app-bot/internal/business/project"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type inspectionDocument struct {
	ID                 string    `firestore:"id"`
	Note               string    `firestore:"note"`
	Status             string    `firestore:"status"`
	SubmittedBy        string    `firestore:"submittedBy"`
	SubmittedAt        time.Time `firestore:"submittedAt"`
	SourceType         string    `firestore:"sourceType"`
	SourceID           string    `firestore:"sourceId"`
	SourceRepositoryID string    `firestore:"sourceRepositoryId"`
	SourceVersion      int64     `firestore:"sourceVersion"`
}

const submissions = "inspectionSubmissions"

func (s *Store) Inspections(ctx context.Context, r feature.Scope, pid, itemID, note string) (result []feature.InspectionSubmission, err error) {
	defer func() { err = grpcstatus.NormalizeContextError(err) }()
	root := s.ownerDoc(r.OwnerID)
	project := root.Collection("projects").Doc(pid)
	item := project.Collection("items").Doc(itemID)
	snap, err := item.Get(ctx)
	if status.Code(err) == codes.NotFound {
		return nil, feature.ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	d, err := decodeItem(snap, pid)
	if err != nil {
		return nil, err
	}
	if d.SourceType != "object" && d.SourceType != "subobject" {
		return nil, feature.ErrInput
	}
	source, err := r.Sources.Authorize(ctx, r.Actor, d.SourceRepositoryID, "")
	if err != nil {
		return nil, err
	}
	collection := item.Collection(submissions)
	newID := collection.NewDoc().ID
	now := time.Now().UTC()
	err = s.Client.RunTransaction(ctx, func(_ context.Context, tx *db.Transaction) error {
		result = []feature.InspectionSubmission{}
		if err := s.checkProjectOwner(tx, r, root); err != nil {
			return err
		}
		snap, err := tx.Get(project)
		if _, err = decodeProject(snap, err, r); err != nil {
			return err
		}
		snap, err = tx.Get(item)
		if status.Code(err) == codes.NotFound {
			return feature.ErrNotFound
		}
		if err != nil {
			return err
		}
		current, err := decodeItem(snap, pid)
		if err != nil {
			return err
		}
		if current.SourceRepositoryID != source.ID {
			return feature.ErrForbidden
		}
		if err = requireProjectLink(tx, project, source.ID); err != nil {
			if errors.Is(err, feature.ErrSourceUnavailable) {
				return feature.ErrForbidden
			}
			return err
		}
		if note == "" {
			if err := s.checkSourceRepository(tx, source); err != nil {
				return err
			}
			rows, err := tx.Documents(collection.OrderBy("submittedAt", db.Desc)).GetAll()
			if err != nil {
				return err
			}
			for _, row := range rows {
				var stored inspectionDocument
				if err := row.DataTo(&stored); err != nil {
					return err
				}
				result = append(result, feature.InspectionSubmission(stored))
			}
			return nil
		}
		ref, err := s.sourceReference(tx, source, itemReference(current))
		if err != nil {
			return err
		}
		if latestID, exists := snap.Data()["latestInspectionId"]; exists {
			id, ok := latestID.(string)
			if !ok || id == "" {
				return feature.ErrUnavailable
			}
			last, err := tx.Get(collection.Doc(id))
			if err != nil {
				return err
			}
			var stored inspectionDocument
			if err = last.DataTo(&stored); err != nil {
				return err
			}
			if stored.Status == "pending" {
				if stored.Note != note || stored.SubmittedBy != r.Actor {
					return feature.ErrInspectionPending
				}
				result = []feature.InspectionSubmission{feature.InspectionSubmission(stored)}
				return nil
			}
		}
		if ref.SourceVersion < 1 {
			return feature.ErrSourceUnavailable
		}
		entry := feature.InspectionSubmission{ID: newID, Note: note, Status: "pending", SubmittedBy: r.Actor, SubmittedAt: now, SourceRepositoryID: source.ID, SourceType: ref.SourceType, SourceID: ref.SourceID, SourceVersion: ref.SourceVersion}
		result = []feature.InspectionSubmission{entry}
		// Writing the reference serializes competing submissions and reference removal.
		if err := tx.Update(item, []db.Update{{Path: "latestInspectionId", Value: newID}}); err != nil {
			return err
		}
		return tx.Create(collection.Doc(newID), inspectionDocument(entry))
	})
	if err != nil {
		return nil, err
	}
	return result, nil
}
