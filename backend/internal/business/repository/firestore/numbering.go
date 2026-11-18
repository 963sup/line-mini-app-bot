package firestore

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"strconv"
	"time"

	db "cloud.google.com/go/firestore"
	"github.com/963sup/line-mini-app-bot/internal/business/repository"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type numberDocument struct {
	SchemaVersion int64     `firestore:"schemaVersion"`
	RepositoryID  string    `firestore:"repositoryId"`
	ResourceType  string    `firestore:"resourceType"`
	ResourceID    string    `firestore:"resourceId"`
	Number        int64     `firestore:"number"`
	CreatedAt     time.Time `firestore:"createdAt"`
}

func (d numberDocument) model() repository.NumberedResource {
	return repository.NumberedResource{RepositoryID: d.RepositoryID, ResourceType: d.ResourceType, ResourceID: d.ResourceID, Number: d.Number, CreatedAt: d.CreatedAt}
}

func (d numberDocument) valid(repositoryID string, number int64) bool {
	return d.SchemaVersion == 1 && d.RepositoryID == repositoryID && d.Number == number && d.model().Valid()
}

type numberCounter struct {
	SchemaVersion int64 `firestore:"schemaVersion"`
	LastNumber    int64 `firestore:"lastNumber"`
}

type numberReceipt struct {
	SchemaVersion int64          `firestore:"schemaVersion"`
	PayloadHash   string         `firestore:"payloadHash"`
	Resource      numberDocument `firestore:"resource"`
}

// CreateNumberedResource commits the shared repository number, lookup, creation
// receipt and caller's business writes together. It is NOT a public allocator.
// The caller must first authenticate ActorID and authorize the domain operation.
// GroupID must come from that authorized repository, never from client claims.
//
// create must persist the business resource using this transaction and the supplied
// ID/number, returning any validation/write error. It may read before its own
// writes. It must not commit, start another transaction, or perform external side
// effects: Firestore can invoke it multiple times. Replays skip create entirely.
// A nil method error confirms commit. Errors return zero values; an ambiguous
// commit error must be retried with the same key to discover the stored result.
func (s *Store) CreateNumberedResource(ctx context.Context, command repository.NumberedCreation, create func(context.Context, *db.Transaction, repository.NumberedResource) error) (repository.NumberedResource, error) {
	if !command.Valid() || create == nil {
		return repository.NumberedResource{}, repository.ErrNumberedCreation
	}
	if s == nil || s.Client == nil || s.ChannelID == "" {
		return repository.NumberedResource{}, repository.ErrUnavailable
	}
	var random [16]byte
	if _, err := rand.Read(random[:]); err != nil {
		return repository.NumberedResource{}, err
	}
	// Allocate stable identity/time outside the retried callback, not a number.
	resourceID, now := hex.EncodeToString(random[:]), time.Now().UTC().Truncate(time.Microsecond)
	parent := s.channel().Collection("repositories").Doc(command.RepositoryID)
	counter := parent.Collection("counters").Doc("resources")
	scope, err := json.Marshal([]string{"numbered-resource-create", command.ActorID, command.Key})
	if err != nil {
		return repository.NumberedResource{}, err
	}
	receipt := parent.Collection("resourceCreationReceipts").Doc(key(string(scope)))
	var result repository.NumberedResource
	err = s.Client.RunTransaction(ctx, func(ctx context.Context, tx *db.Transaction) error {
		result = repository.NumberedResource{}
		container, err := decode(tx.Get(parent))
		if err != nil {
			return err
		}
		if container.GroupID != command.GroupID || container.Status != "ACTIVE" {
			return repository.ErrNotFound
		}
		prior, err := tx.Get(receipt)
		if err == nil {
			var stored numberReceipt
			if err := prior.DataTo(&stored); err != nil {
				return err
			}
			if stored.SchemaVersion != 1 || !stored.Resource.valid(command.RepositoryID, stored.Resource.Number) {
				return repository.ErrUnavailable
			}
			if stored.PayloadHash != command.PayloadHash || stored.Resource.ResourceType != command.ResourceType {
				return repository.ErrCreationConflict
			}
			lookup, err := tx.Get(parent.Collection("resourceNumbers").Doc(strconv.FormatInt(stored.Resource.Number, 10)))
			if err != nil {
				return repository.ErrUnavailable
			}
			var indexed numberDocument
			if err := lookup.DataTo(&indexed); err != nil {
				return err
			}
			if indexed != stored.Resource {
				return repository.ErrUnavailable
			}
			result = stored.Resource.model()
			return nil
		}
		if status.Code(err) != codes.NotFound {
			return err
		}
		current, err := tx.Get(counter)
		var state numberCounter
		if err == nil {
			if err := current.DataTo(&state); err != nil {
				return err
			}
			if state.SchemaVersion != 1 || state.LastNumber < 1 || state.LastNumber > repository.MaxResourceNumber {
				return repository.ErrUnavailable
			}
		} else if status.Code(err) != codes.NotFound {
			return err
		} else {
			state.SchemaVersion = 1
		}
		if state.LastNumber == repository.MaxResourceNumber {
			return repository.ErrNumberExhausted
		}
		state.LastNumber++
		index := parent.Collection("resourceNumbers").Doc(strconv.FormatInt(state.LastNumber, 10))
		if _, err := tx.Get(index); status.Code(err) != codes.NotFound {
			if err != nil {
				return err
			}
			// Never overwrite an existing reference if a counter was reset/corrupt.
			return repository.ErrUnavailable
		}
		doc := numberDocument{SchemaVersion: 1, RepositoryID: command.RepositoryID, ResourceType: command.ResourceType, ResourceID: resourceID, Number: state.LastNumber, CreatedAt: now}
		if err := create(ctx, tx, doc.model()); err != nil {
			return err
		}
		// All common reads precede the caller's writes; no reads follow create.
		if err := tx.Set(counter, state); err != nil {
			return err
		}
		if err := tx.Create(index, doc); err != nil {
			return err
		}
		if err := tx.Create(receipt, numberReceipt{SchemaVersion: 1, PayloadHash: command.PayloadHash, Resource: doc}); err != nil {
			return err
		}
		result = doc.model()
		return nil
	})
	if err != nil {
		return repository.NumberedResource{}, err
	}
	return result, nil
}

// GetResourceNumber is a persistence lookup, not an authorization check. A future
// domain use case must authenticate and authorize before using or exposing it.
func (s *Store) GetResourceNumber(ctx context.Context, repositoryID string, number int64) (repository.NumberedResource, error) {
	if !repository.ValidID(repositoryID) || number < 1 || number > repository.MaxResourceNumber {
		return repository.NumberedResource{}, repository.ErrNotFound
	}
	if s == nil || s.Client == nil || s.ChannelID == "" {
		return repository.NumberedResource{}, repository.ErrUnavailable
	}
	ref := s.channel().Collection("repositories").Doc(repositoryID).Collection("resourceNumbers").Doc(strconv.FormatInt(number, 10))
	snap, err := ref.Get(ctx)
	if status.Code(err) == codes.NotFound {
		return repository.NumberedResource{}, repository.ErrNotFound
	}
	if err != nil {
		return repository.NumberedResource{}, err
	}
	var doc numberDocument
	if err := snap.DataTo(&doc); err != nil {
		return repository.NumberedResource{}, err
	}
	if !doc.valid(repositoryID, number) {
		return repository.NumberedResource{}, repository.ErrUnavailable
	}
	return doc.model(), nil
}
