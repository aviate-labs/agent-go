package identity

import (
	"github.com/niccolofant/agent-go/principal"
)

// Identity is an identity that can sign messages.
type Identity interface {
	// Sender returns the principal of the identity.
	Sender() principal.Principal
	// Sign signs the given message.
	Sign(msg []byte) ([]byte, error)
	// PublicKey returns the public key of the identity.
	PublicKey() []byte
	// Verify verifies the signature of the given message.
	Verify(msg, sig []byte) bool
	// ToPEM returns the PEM representation of the identity.
	ToPEM() ([]byte, error)
}
