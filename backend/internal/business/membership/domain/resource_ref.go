// Package domain holds Membership rules independent of storage and transport.
package domain

import (
	"errors"
	"strings"
)

var ErrInvalidIdentity = errors.New("invalid membership identity")

// ResourceRef identifies a resource within its storage channel. It is
// comparable by value and carries no authorization or provider credentials.
type ResourceRef struct {
	channelID string
	kind      string
	id        string
}

func NewResourceRef(channelID, kind, id string) (ResourceRef, error) {
	if !validID(channelID) || !validID(kind) || !validID(id) {
		return ResourceRef{}, ErrInvalidIdentity
	}
	return ResourceRef{channelID: channelID, kind: kind, id: id}, nil
}

func (r ResourceRef) ChannelID() string { return r.channelID }
func (r ResourceRef) Kind() string      { return r.kind }
func (r ResourceRef) ID() string        { return r.id }

func validID(value string) bool {
	return value != "" && strings.TrimSpace(value) == value
}
