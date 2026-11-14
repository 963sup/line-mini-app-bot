package firestore

import (
	"time"

	feature "github.com/963sup/line-mini-app-bot/internal/business/objectimport"
)

// v1 shapes stay independent of domain field additions/renames.
type storedEvidence struct {
	Page  int    `json:"page"`
	Quote string `json:"quote"`
}
type storedRow struct {
	ID             string           `json:"id"`
	Name           string           `json:"name"`
	Description    string           `json:"description"`
	Type           string           `json:"type"`
	Quantity       string           `json:"quantity"`
	Unit           string           `json:"unit"`
	MaterialCode   string           `json:"materialCode"`
	SourceItemNo   string           `json:"sourceItemNo"`
	Evidence       []storedEvidence `json:"evidence"`
	ResultObjectID string           `json:"resultObjectId"`
}
type storedJob struct {
	ID                         string      `json:"id"`
	RepositoryID               string      `json:"repositoryId"`
	CreatorID                  string      `json:"creatorId"`
	Status                     string      `json:"status"`
	Version                    int64       `json:"version"`
	ConfirmedFromReviewVersion int64       `json:"confirmedFromReviewVersion"`
	SourceDocument             string      `json:"sourceDocument"`
	SourceRevision             string      `json:"sourceRevision"`
	SourceHash                 string      `json:"sourceHash"`
	Rows                       []storedRow `json:"rows"`
	CreatedAt                  time.Time   `json:"createdAt"`
	UpdatedAt                  time.Time   `json:"updatedAt"`
	ReviewedAt                 *time.Time  `json:"reviewedAt,omitempty"`
}

func toStored(j feature.Job) storedJob {
	rows := make([]storedRow, 0, len(j.Rows))
	for _, r := range j.Rows {
		proof := make([]storedEvidence, 0, len(r.Evidence))
		for _, e := range r.Evidence {
			proof = append(proof, storedEvidence{Page: e.Page, Quote: e.Quote})
		}
		rows = append(rows, storedRow{ID: r.ID, Name: r.Name, Description: r.Description, Type: r.Type, Quantity: r.Quantity, Unit: r.Unit, MaterialCode: r.MaterialCode, SourceItemNo: r.SourceItemNo, Evidence: proof, ResultObjectID: r.ResultObjectID})
	}
	return storedJob{ID: j.ID, RepositoryID: j.RepositoryID, CreatorID: j.CreatorID, Status: j.Status, Version: j.Version, ConfirmedFromReviewVersion: j.ConfirmedFromReviewVersion, SourceDocument: j.SourceDocument, SourceRevision: j.SourceRevision, SourceHash: j.SourceHash, Rows: rows, CreatedAt: j.CreatedAt, UpdatedAt: j.UpdatedAt, ReviewedAt: j.ReviewedAt}
}

func fromStored(j storedJob) feature.Job {
	rows := make([]feature.Row, 0, len(j.Rows))
	for _, r := range j.Rows {
		proof := make([]feature.Evidence, 0, len(r.Evidence))
		for _, e := range r.Evidence {
			proof = append(proof, feature.Evidence{Page: e.Page, Quote: e.Quote})
		}
		rows = append(rows, feature.Row{ID: r.ID, Fields: feature.Fields{Name: r.Name, Description: r.Description, Type: r.Type, Quantity: r.Quantity, Unit: r.Unit, MaterialCode: r.MaterialCode}, SourceItemNo: r.SourceItemNo, Evidence: proof, ResultObjectID: r.ResultObjectID})
	}
	return feature.Job{ID: j.ID, RepositoryID: j.RepositoryID, CreatorID: j.CreatorID, Status: j.Status, Version: j.Version, ConfirmedFromReviewVersion: j.ConfirmedFromReviewVersion, SourceDocument: j.SourceDocument, SourceRevision: j.SourceRevision, SourceHash: j.SourceHash, Rows: rows, CreatedAt: j.CreatedAt, UpdatedAt: j.UpdatedAt, ReviewedAt: j.ReviewedAt}
}
