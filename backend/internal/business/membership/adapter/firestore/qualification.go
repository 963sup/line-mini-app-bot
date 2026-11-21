package membershipdb

import (
	"cloud.google.com/go/firestore"
	"github.com/963sup/line-mini-app-bot/internal/business/membership/domain"
	"github.com/963sup/line-mini-app-bot/internal/platform/firestorekey"
	"github.com/google/uuid"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// OrganizationMember reads the independent authority within the caller's
// transaction. Missing or inactive qualifications never fall back to legacy.
func OrganizationMember(tx *firestore.Transaction, root *firestore.DocumentRef, organizationID, actor string) (bool, error) {
	q, err := parentQualification(tx, root, "organization", organizationID, actor)
	return q.ActiveFor("organization"), err
}

// TeamMember checks the team relationship and the exact generation of the
// qualification from which it derives. Team supplies the management scope
// from the same transaction.
func TeamMember(tx *firestore.Transaction, root *firestore.DocumentRef, teamID, parentKind, parentID, actor string) (bool, error) {
	q, err := readQualification(tx, root, "team", teamID, actor)
	if err != nil || !q.ActiveFor("team") || q.ParentKind != parentKind || q.ParentID != parentID {
		return false, err
	}
	p, err := parentQualification(tx, root, parentKind, parentID, actor)
	return q.MatchesParent(parentKind, parentID, p), err
}

func parentQualification(tx *firestore.Transaction, root *firestore.DocumentRef, kind, id, actor string) (domain.Qualification, error) {
	if (kind != "organization" && kind != "enterprise") || len(id) != 36 || uuid.Validate(id) != nil {
		return domain.Qualification{}, nil
	}
	doc, err := tx.Get(root.Collection(kind + "s").Doc(id))
	if status.Code(err) == codes.NotFound {
		return domain.Qualification{}, nil
	}
	if err != nil {
		return domain.Qualification{}, err
	}
	var resource struct {
		Mode     string `firestore:"membershipMode"`
		Disabled bool   `firestore:"disabled"`
	}
	if err := doc.DataTo(&resource); err != nil {
		return domain.Qualification{}, err
	}
	if resource.Mode != "membership" || resource.Disabled {
		return domain.Qualification{}, nil
	}
	return readQualification(tx, root, kind, id, actor)
}

func readQualification(tx *firestore.Transaction, root *firestore.DocumentRef, kind, id, actor string) (domain.Qualification, error) {
	m, err := readRelationship(tx, root, kind, id, actor)
	return m.Qualification, err
}

func relationshipRef(root *firestore.DocumentRef, kind, id, actor string) *firestore.DocumentRef {
	return root.Collection("memberships").Doc(firestorekey.Encode(kind, id, actor))
}

type relationshipRecord struct {
	Status           string `firestore:"status"`
	Role             string `firestore:"role"`
	Generation       int64  `firestore:"generation"`
	ParentKind       string `firestore:"parentKind"`
	ParentID         string `firestore:"parentId"`
	ParentGeneration int64  `firestore:"parentGeneration"`
	Version          int64  `firestore:"version"`
}

func readRelationship(tx *firestore.Transaction, root *firestore.DocumentRef, kind, id, actor string) (domain.TeamMembership, error) {
	if actor == "" || id == "" {
		return domain.TeamMembership{}, nil
	}
	doc, err := tx.Get(relationshipRef(root, kind, id, actor))
	if status.Code(err) == codes.NotFound {
		return domain.TeamMembership{}, nil
	}
	if err != nil {
		return domain.TeamMembership{}, err
	}
	var r relationshipRecord
	if err := doc.DataTo(&r); err != nil {
		return domain.TeamMembership{}, err
	}
	return domain.TeamMembership{Version: r.Version, Qualification: domain.Qualification{
		Status: r.Status, Role: r.Role, Generation: r.Generation,
		ParentKind: r.ParentKind, ParentID: r.ParentID, ParentGeneration: r.ParentGeneration,
	}}, nil
}
