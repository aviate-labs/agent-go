package identity

import (
	"github.com/aviate-labs/agent-go/principal"
)

type AnonymousIdentity struct{}

func (id AnonymousIdentity) PublicKey() []byte {
	return nil
}

func (id AnonymousIdentity) Sender() principal.Principal {
	return principal.AnonymousID
}

func (id AnonymousIdentity) Sign(msg []byte) []byte {
	return nil
}
