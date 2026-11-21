package domain

// Opening binds a resource to its initial owner before persistence. The caller
// must supply the actor established by the authenticated entry point.
type Opening struct {
	resource ResourceRef
	ownerID  string
}

func OpenScope(resource ResourceRef, actorID string) (Opening, error) {
	if resource == (ResourceRef{}) || !validID(actorID) {
		return Opening{}, ErrInvalidIdentity
	}
	return Opening{resource: resource, ownerID: actorID}, nil
}

func (o Opening) Resource() ResourceRef { return o.resource }
func (o Opening) OwnerID() string       { return o.ownerID }
