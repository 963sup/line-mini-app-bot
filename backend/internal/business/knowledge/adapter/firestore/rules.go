// Package firestore reads the current published knowledge for a template.
package firestore

import (
	"context"
	"errors"
	"slices"
	"strings"

	db "cloud.google.com/go/firestore"
	org "github.com/963sup/line-mini-app-bot/internal/business/internal/adapter/firestore"
	"github.com/963sup/line-mini-app-bot/internal/business/knowledge/application/port"
	"github.com/963sup/line-mini-app-bot/internal/platform/firestore/key"
	"github.com/google/uuid"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type Reader struct {
	DB      *db.Client
	Channel string
}

var _ port.Reader = Reader{}

func (r Reader) Retrieve(ctx context.Context, q port.Lookup) ([]port.PublishedRule, error) {
	if q.Actor == "" || uuid.Validate(q.Organization) != nil || strings.TrimSpace(q.TemplateVersion) == "" || len(q.TemplateVersion) > 200 {
		return nil, errors.New("invalid knowledge scope")
	}
	result, err := org.Run(ctx, r.DB, r.Channel, func(tx *db.Transaction, root *db.DocumentRef) (*[]port.PublishedRule, error) {
		organization := root.Collection("organizations").Doc(q.Organization)
		membership, err := org.Read(tx, organization)
		if err != nil {
			return nil, err
		}
		if membership.Owner != q.Actor && !slices.Contains(membership.Members, q.Actor) {
			return nil, errors.New("knowledge scope unavailable")
		}
		rules, err := publishedRules(tx, organization, q.TemplateVersion)
		return &rules, err
	})
	if err != nil {
		return nil, err
	}
	return *result, nil
}

func publishedRules(tx *db.Transaction, organization *db.DocumentRef, templateVersion string) ([]port.PublishedRule, error) {
	// Publication replaces this projection; withdrawal deletes it. Drafts live elsewhere.
	snap, err := tx.Get(organization.Collection("publishedKnowledge").Doc(firestorekey.Digest([]byte(templateVersion))))
	rules := []port.PublishedRule{}
	if status.Code(err) == codes.NotFound {
		return rules, nil
	}
	if err != nil {
		return nil, err
	}
	var projection struct {
		TemplateVersion string               `firestore:"templateVersion"`
		Rules           []port.PublishedRule `firestore:"rules"`
	}
	if err := snap.DataTo(&projection); err != nil {
		return nil, err
	}
	if projection.TemplateVersion != templateVersion {
		return nil, errors.New("knowledge version mismatch")
	}
	for _, rule := range projection.Rules {
		if rule.ID == "" || rule.Version == "" || rule.Source == "" || rule.Text == "" {
			return nil, errors.New("incomplete published rule")
		}
	}
	return projection.Rules, nil
}
