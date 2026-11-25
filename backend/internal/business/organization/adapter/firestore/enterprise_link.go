package organizationdb

import (
	"cloud.google.com/go/firestore"
	dto "github.com/963sup/line-mini-app-bot/internal/business/enterprise/application/dto"
	orgstore "github.com/963sup/line-mini-app-bot/internal/business/internal/adapter/firestore"
)

// AttachEnterprise joins the enterprise owner's transaction.
func AttachEnterprise(tx *firestore.Transaction, root *firestore.DocumentRef, actor, organizationID, enterpriseID string) error {
	ref := root.Collection("organizations").Doc(organizationID)
	doc, err := orgstore.Read(tx, ref)
	if err != nil {
		return err
	}
	org, err := doc.LegacyOrganization().AttachEnterprise(actor, enterpriseID)
	if err != nil || doc.Enterprise == org.EnterpriseID {
		return err
	}
	return tx.Update(ref, []firestore.Update{{Path: "enterpriseId", Value: org.EnterpriseID}})
}

func (s Store) EnterpriseOrganizations(tx *firestore.Transaction, root *firestore.DocumentRef, actor, enterpriseID string, admin bool) ([]dto.Link, []dto.Link, error) {
	var visible, assignable []dto.Link
	docs, e := tx.Documents(root.Collection("organizations").Where("enterpriseId", "==", enterpriseID)).GetAll()
	if e != nil {
		return nil, nil, e
	}
	for _, d := range docs {
		var org orgstore.Document
		if e := d.DataTo(&org); e != nil {
			return nil, nil, e
		}
		canView := admin || org.VisibleLegacyOrganization(actor)
		if !canView && org.MembershipMode == "membership" && !org.Disabled && s.OrganizationMember != nil {
			canView, e = s.OrganizationMember(tx, root, d.Ref.ID, actor)
			if e != nil {
				return nil, nil, e
			}
		}
		if admin && org.LegacyOrganizationEnabled() {
			assignable = append(assignable, dto.Link{ID: d.Ref.ID, Name: org.Name})
		}
		if canView {
			visible = append(visible, dto.Link{ID: d.Ref.ID, Name: org.Name})
		}
	}

	return visible, assignable, nil
}
