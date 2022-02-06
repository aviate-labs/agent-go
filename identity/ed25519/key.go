package ed25519

import (
	"crypto/ed25519"
)

type Identity struct {
	privateKey ed25519.PrivateKey
}

func (id Identity) Sign(msg []byte) ([]byte, error) {
	return ed25519.Sign(id.privateKey, msg), nil
}
