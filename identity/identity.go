package identity

import (
	"crypto/elliptic"
	"github.com/aviate-labs/agent-go/principal"
	"math/big"
)

func marshal(curve elliptic.Curve, x, y *big.Int) []byte {
	byteLen := (curve.Params().BitSize + 7) / 8
	ret := make([]byte, 1+2*byteLen)
	ret[0] = 4 // uncompressed point
	x.FillBytes(ret[1 : 1+byteLen])
	y.FillBytes(ret[1+byteLen : 1+2*byteLen])
	return ret
}

// Identity is an identity that can sign messages.
type Identity interface {
	// Sender returns the principal of the identity.
	Sender() principal.Principal
	// Sign signs the given message.
	Sign(msg []byte) []byte
	// PublicKey returns the public key of the identity.
	PublicKey() []byte
}
