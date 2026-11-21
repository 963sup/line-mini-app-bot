package entity

// Reference retains source identity only; display fields are resolved for each read.
type Reference struct {
	ItemStatus                                                        string
	SourceVersion                                                     int64
	ID, SourceType, SourceID, SourceRepositoryID                      string
	SourceUnavailable                                                 bool
	SourceTitle, SourceState, Quantity, Unit, CompletedQuantity, Type string
}

func (r Reference) SupportsInspection() bool {
	return r.SourceType == "object"
}

func ValidReference(r Reference) bool {
	if !ValidID(r.SourceID) {
		return false
	}
	switch r.SourceType {
	case "object", "issue", "discussion":
		return true
	default:
		return false
	}
}
