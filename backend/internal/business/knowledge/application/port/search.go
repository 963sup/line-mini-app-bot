package port

import "context"

// EvidenceSearch uses a vector from the same embedding space as the index.
// Actor and Organization come from the calling use case's trusted scope.
type EvidenceSearch struct {
	Actor, Organization string
	Vector              []float64
	Limit               int
}

// EvidenceSearcher returns currently published rules in nearest-first order.
// Revoked or superseded index entries can make the result shorter than Limit.
type EvidenceSearcher interface {
	SearchEvidence(context.Context, EvidenceSearch) ([]PublishedRule, error)
}
