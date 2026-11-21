package entity

// TransferOwnership determines the new owner from the current ownership and
// membership facts. The caller must load and save them in one transaction.
func (o Organization) TransferOwnership(actor, successor string, successorIsMember bool) (Organization, error) {
	if actor == "" || o.OwnerUserID != actor {
		return Organization{}, ErrNotFound
	}
	if successor == "" || (successor != actor && !successorIsMember) {
		return Organization{}, ErrInvalid
	}
	o.OwnerUserID = successor
	return o, nil
}
