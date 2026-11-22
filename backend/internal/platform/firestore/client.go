// Package firestore provides shared Firestore client setup.
package firestore

import (
	"context"

	"cloud.google.com/go/firestore"
)

// NewClient opens the requested database. The caller owns the client and closes it.
func NewClient(ctx context.Context, projectID, databaseID string) (*firestore.Client, error) {
	if databaseID == "" {
		databaseID = "(default)"
	}
	return firestore.NewClientWithDatabase(ctx, projectID, databaseID)
}
