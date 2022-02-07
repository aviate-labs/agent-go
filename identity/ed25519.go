package identity

import (
	"crypto/ed25519"
	"crypto/x509/pkix"
	"encoding/asn1"

	"github.com/aviate-labs/principal-go"
)

func derEncodePublicKey(key ed25519.PublicKey) ([]byte, error) {
	return asn1.Marshal(struct {
		Algorithm pkix.AlgorithmIdentifier
		PublicKey asn1.BitString
	}{
		Algorithm: pkix.AlgorithmIdentifier{
			Algorithm: asn1.ObjectIdentifier{
				3, 3, 101, 112,
			},
		},
		PublicKey: asn1.BitString{
			BitLength: len(key) * 8,
			Bytes:     key,
		},
	})
}

type Ed25519Identity struct {
	privateKey ed25519.PrivateKey
}

func (id Ed25519Identity) PublicKey() []byte {
	der, _ := derEncodePublicKey(id.privateKey.Public().(ed25519.PublicKey))
	return der
}

func (id Ed25519Identity) Sender() principal.Principal {
	return principal.NewSelfAuthenticating(id.PublicKey())
}

func (id Ed25519Identity) Sign(data []byte) []byte {
	return ed25519.Sign(id.privateKey, data)
}
