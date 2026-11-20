package port

import "context"

type Lookup struct {
	Actor, Organization, TemplateVersion string
}

type PublishedRule struct {
	ID, Version, Source, Text string
}

// Reader returns only published rules visible to the actor for this template.
type Reader interface {
	Retrieve(context.Context, Lookup) ([]PublishedRule, error)
}
