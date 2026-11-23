package event

// QuantityChanged describes the quantity fields of a committed object revision.
type QuantityChanged struct{ BeforeQuantity, BeforeUnit, AfterQuantity, AfterUnit string }

func QuantityChange(before, beforeUnit, after, afterUnit string) (QuantityChanged, bool) {
	if before == after {
		return QuantityChanged{}, false
	}
	return QuantityChanged{before, beforeUnit, after, afterUnit}, true
}

// Revision identifies a quantity change without disclosing actor or document text.
type Revision struct {
	ChannelID    string
	RepositoryID string
	ObjectID     string
	Version      int64
	Change       QuantityChanged
}

type PendingRevision struct {
	ID       string
	Revision Revision
}
