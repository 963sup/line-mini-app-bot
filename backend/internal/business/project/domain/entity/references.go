package entity

// Reference retains source identity only; display fields are resolved for each read.
type Reference struct {
	ItemStatus                                                        string
	SourceVersion                                                     int64
	ID, SourceType, SourceID, ParentObjectID, SourceRepositoryID      string
	SourceUnavailable                                                 bool
	SourceTitle, SourceState, Quantity, Unit, CompletedQuantity, Type string
}

func ValidReference(r Reference) bool {
	if !ValidID(r.SourceID) {
		return false
	}
	switch r.SourceType {
	case "subobject":
		return ValidID(r.ParentObjectID)
	case "object", "issue", "discussion":
		return r.ParentObjectID == ""
	default:
		return false
	}
}
