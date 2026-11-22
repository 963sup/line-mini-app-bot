package firestore

import (
	"context"
	"errors"
	"math"
	"slices"
	"strings"

	db "cloud.google.com/go/firestore"
	org "github.com/963sup/line-mini-app-bot/internal/business/internal/adapter/firestore"
	"github.com/963sup/line-mini-app-bot/internal/business/knowledge/application/port"
	"github.com/google/uuid"
)

var _ port.EvidenceSearcher = Reader{}

func (r Reader) SearchEvidence(ctx context.Context, q port.EvidenceSearch) ([]port.PublishedRule, error) {
	if q.Actor == "" || uuid.Validate(q.Organization) != nil || q.Limit < 1 || q.Limit > 20 || len(q.Vector) < 1 || len(q.Vector) > 2048 {
		return nil, errors.New("invalid knowledge search")
	}
	for _, v := range q.Vector {
		if math.IsNaN(v) || math.IsInf(v, 0) {
			return nil, errors.New("invalid knowledge search vector")
		}
	}
	result, err := org.Run(ctx, r.DB, r.Channel, func(tx *db.Transaction, root *db.DocumentRef) (*[]port.PublishedRule, error) {
		organization := root.Collection("organizations").Doc(q.Organization)
		membership, err := org.Read(tx, organization)
		if err != nil {
			return nil, err
		}
		if membership.Disabled || (membership.Owner != q.Actor && !slices.Contains(membership.Members, q.Actor)) {
			return nil, errors.New("knowledge scope unavailable")
		}
		query := organization.Collection("publishedKnowledgeVectors").
			FindNearest("embedding", db.Vector64(q.Vector), q.Limit, db.DistanceMeasureEuclidean, nil)
		hits, err := tx.Documents(query).GetAll()
		if err != nil {
			return nil, err
		}
		matches := make([]port.PublishedRule, 0, len(hits))
		byTemplate := map[string][]port.PublishedRule{}
		type ruleKey struct{ template, id, version string }
		seen := map[ruleKey]bool{}
		for _, hit := range hits {
			var ref struct {
				TemplateVersion string `firestore:"templateVersion"`
				RuleID          string `firestore:"ruleId"`
				RuleVersion     string `firestore:"ruleVersion"`
			}
			if err := hit.DataTo(&ref); err != nil {
				return nil, err
			}
			if strings.TrimSpace(ref.TemplateVersion) == "" || len(ref.TemplateVersion) > 200 || ref.RuleID == "" || ref.RuleVersion == "" {
				continue
			}
			key := ruleKey{ref.TemplateVersion, ref.RuleID, ref.RuleVersion}
			if seen[key] {
				continue
			}
			seen[key] = true
			published, ok := byTemplate[ref.TemplateVersion]
			if !ok {
				published, err = publishedRules(tx, organization, ref.TemplateVersion)
				if err != nil {
					return nil, err
				}
				byTemplate[ref.TemplateVersion] = published
			}
			for _, rule := range published {
				if rule.ID == ref.RuleID && rule.Version == ref.RuleVersion {
					matches = append(matches, rule)
					break
				}
			}
		}
		return &matches, nil
	})
	if err != nil {
		return nil, err
	}
	return *result, nil
}
