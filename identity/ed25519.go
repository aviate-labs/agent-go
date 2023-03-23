package identity

import (
	"crypto/ed25519"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/asn1"
	"encoding/pem"
	"fmt"

	"github.com/aviate-labs/agent-go/principal"
)

func derEncodePublicKey(key ed25519.PublicKey) ([]byte, error) {
	return asn1.Marshal(struct {
		Algorithm pkix.AlgorithmIdentifier
		PublicKey asn1.BitString
	}{
		Algorithm: pkix.AlgorithmIdentifier{
			Algorithm: asn1.ObjectIdentifier{
				1, 3, 101, 112,
			},
		},
		PublicKey: asn1.BitString{
			BitLength: len(key) * 8,
			Bytes:     key,
		},
	})
}

// Ed25519Identity is an identity based on an Ed25519 key pair.
type Ed25519Identity struct {
	privateKey ed25519.PrivateKey
	publicKey  ed25519.PublicKey
}

// NewEd25519Identity creates a new identity based on the given key pair.
func NewEd25519Identity(publicKey ed25519.PublicKey, privateKey ed25519.PrivateKey) Ed25519Identity {
	return Ed25519Identity{
		publicKey:  publicKey,
		privateKey: privateKey,
	}
}

// NewEd25519IdentityFromPEM creates a new identity from the given PEM file.
func NewEd25519IdentityFromPEM(data []byte) (*Ed25519Identity, error) {
	block, _ := pem.Decode(data)
	if block.Type != "PRIVATE KEY" {
		return nil, fmt.Errorf("invalid pem file")
	}
	privateKey, err := x509.ParsePKCS8PrivateKey(block.Bytes)
	if err != nil {
		return nil, err
	}
	switch k := privateKey.(type) {
	case ed25519.PrivateKey:
		return &Ed25519Identity{
			privateKey: k,
			publicKey:  k.Public().(ed25519.PublicKey),
		}, nil
	default:
		return nil, fmt.Errorf("unknown key type")
	}
}

// PublicKey returns the public key of the identity.
func (id Ed25519Identity) PublicKey() []byte {
	der, _ := derEncodePublicKey(id.publicKey)
	return der
}

// Sender returns the principal of the identity.
func (id Ed25519Identity) Sender() principal.Principal {
	return principal.NewSelfAuthenticating(id.PublicKey())
}

// Sign signs the given message.
func (id Ed25519Identity) Sign(data []byte) []byte {
	return ed25519.Sign(id.privateKey, data)
}

// ToPEM returns the PEM representation of the identity.
func (id Ed25519Identity) ToPEM() ([]byte, error) {
	data, err := x509.MarshalPKCS8PrivateKey(id.privateKey)
	if err != nil {
		return nil, err
	}
	return pem.EncodeToMemory(&pem.Block{
		Type:  "PRIVATE KEY",
		Bytes: data,
	}), nil
}
