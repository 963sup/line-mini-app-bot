package richmenu

import (
	"context"
	"crypto/rand"
	"errors"
	"time"

	"cloud.google.com/go/firestore"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

var errNavigationBusy = errors.New("another LINE navigation publication is active")

// AcquireLease serializes menu publication for one LINE channel.
func AcquireLease(ctx context.Context, db *firestore.Client, channel string) (func() error, error) {
	ref := db.Collection("line_navigation_publication_leases").Doc(channel)
	owner := rand.Text()
	for {
		err := db.RunTransaction(ctx, func(ctx context.Context, tx *firestore.Transaction) error {
			doc, err := tx.Get(ref)
			if err != nil && status.Code(err) != codes.NotFound {
				return err
			}
			if err == nil {
				until, ok := doc.Data()["expiresAt"].(time.Time)
				if !ok || until.IsZero() {
					return errors.New("invalid LINE navigation lease; refusing publication")
				}
				if until.After(time.Now()) {
					return errNavigationBusy
				}
			}
			return tx.Set(ref, map[string]any{"owner": owner, "expiresAt": time.Now().Add(10 * time.Minute)})
		})
		if err == nil {
			break
		}
		if ctx.Err() != nil {
			return nil, ctx.Err()
		}
		if !errors.Is(err, errNavigationBusy) {
			return nil, err
		}
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		case <-time.After(5 * time.Second):
		}
	}
	return func() error {
		releaseCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		return db.RunTransaction(releaseCtx, func(ctx context.Context, tx *firestore.Transaction) error {
			doc, err := tx.Get(ref)
			if status.Code(err) == codes.NotFound {
				return nil
			}
			if err != nil {
				return err
			}
			if doc.Data()["owner"] != owner {
				return nil
			}
			return tx.Delete(ref)
		})
	}, nil
}
