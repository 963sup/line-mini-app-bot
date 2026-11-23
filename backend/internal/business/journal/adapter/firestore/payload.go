package firestore

import (
	"time"

	journalentity "github.com/963sup/line-mini-app-bot/internal/business/journal/domain/entity"
)

// These payloads preserve the stored JSON and receipt hash format, including
// field order, absent optional fields, and the distinction between null and [].
// Domain model changes must not implicitly change existing receipts or records.
type referencePayload struct{ Kind, ID string }

type workPayload struct {
	Name       string
	Percent    int
	Reference  *referencePayload
	ObjectID   string   `json:",omitempty"`
	SpaceIDs   []string `json:",omitempty"`
	ObjectName string   `json:",omitempty"`
	SpaceNames []string `json:",omitempty"`
}

type personnelPayload struct{ Management, Control, Workers []string }

type draftPayload struct {
	GroupID         string
	RepositoryID    string `json:",omitempty"`
	WorkDate        time.Time
	Personnel       personnelPayload
	Today, Tomorrow []workPayload
}

type entryPayload struct {
	draftPayload
	ID, CreatorID string
	CreatedAt     time.Time
}

func workData(rows []journalentity.Work) []workPayload {
	if rows == nil {
		return nil
	}
	result := make([]workPayload, len(rows))
	for i, w := range rows {
		result[i] = workPayload{Name: w.Name, Percent: w.Percent, ObjectID: w.ObjectID, SpaceIDs: w.SpaceIDs, ObjectName: w.ObjectName, SpaceNames: w.SpaceNames}
		if w.Reference != nil {
			result[i].Reference = &referencePayload{Kind: w.Reference.Kind, ID: w.Reference.ID}
		}
	}
	return result
}

func draftData(d journalentity.Draft) draftPayload {
	return draftPayload{
		GroupID: d.GroupID, RepositoryID: d.RepositoryID, WorkDate: d.WorkDate,
		Personnel: personnelPayload(d.Personnel), Today: workData(d.Today), Tomorrow: workData(d.Tomorrow),
	}
}

func entryData(e journalentity.Entry) entryPayload {
	return entryPayload{draftPayload: draftData(e.Draft), ID: e.ID, CreatorID: e.CreatorID, CreatedAt: e.CreatedAt}
}

func (e entryPayload) model() journalentity.Entry {
	return journalentity.Entry{
		Draft: journalentity.Draft{
			GroupID: e.GroupID, RepositoryID: e.RepositoryID, WorkDate: e.WorkDate,
			Personnel: journalentity.Personnel(e.Personnel), Today: workModels(e.Today), Tomorrow: workModels(e.Tomorrow),
		},
		ID: e.ID, CreatorID: e.CreatorID, CreatedAt: e.CreatedAt,
	}
}

func workModels(rows []workPayload) []journalentity.Work {
	if rows == nil {
		return nil
	}
	result := make([]journalentity.Work, len(rows))
	for i, w := range rows {
		result[i] = journalentity.Work{Name: w.Name, Percent: w.Percent, ObjectID: w.ObjectID, SpaceIDs: w.SpaceIDs, ObjectName: w.ObjectName, SpaceNames: w.SpaceNames}
		if w.Reference != nil {
			result[i].Reference = &journalentity.Reference{Kind: w.Reference.Kind, ID: w.Reference.ID}
		}
	}
	return result
}
