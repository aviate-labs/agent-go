package identity

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/sha256"
	"crypto/x509"
	"encoding/asn1"
	"encoding/pem"
	"fmt"
	"math/big"

	"github.com/niccolofant/agent-go/principal"
)

var (
	prime256v1OID = asn1.ObjectIdentifier{1, 2, 840, 10045, 3, 1, 7}
	p256Order     = elliptic.P256().Params().N
	p256HalfOrder = new(big.Int).Rsh(p256Order, 1)
)

func derEncodePrime256v1PublicKey(key *ecdsa.PublicKey) ([]byte, error) {
	pub, err := key.Bytes()
	if err != nil {
		return nil, err
	}
	return asn1.Marshal(ecPublicKey{
		Metadata: []asn1.ObjectIdentifier{
			ecPublicKeyOID,
			prime256v1OID,
		},
		PublicKey: asn1.BitString{
			Bytes: pub,
		},
	})
}

// Prime256v1Identity is an identity based on a P-256 key pair.
type Prime256v1Identity struct {
	privateKey *ecdsa.PrivateKey
	publicKey  *ecdsa.PublicKey
}

// NewPrime256v1Identity creates a new identity based on the given key pair.
func NewPrime256v1Identity(privateKey *ecdsa.PrivateKey) *Prime256v1Identity {
	return &Prime256v1Identity{
		privateKey: privateKey,
		publicKey:  privateKey.Public().(*ecdsa.PublicKey),
	}
}

// NewPrime256v1IdentityFromPEM creates a new identity from the given PEM file.
func NewPrime256v1IdentityFromPEM(data []byte) (*Prime256v1Identity, error) {
	block, remainder := pem.Decode(data)
	if block == nil || block.Type != "EC PRIVATE KEY" || len(remainder) != 0 {
		return nil, fmt.Errorf("invalid pem file")
	}
	privateKey, err := x509.ParseECPrivateKey(block.Bytes)
	if err != nil {
		return nil, err
	}
	return NewPrime256v1Identity(privateKey), nil
}

// NewRandomPrime256v1Identity creates a new identity with a random key pair.
func NewRandomPrime256v1Identity() (*Prime256v1Identity, error) {
	privateKey, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		return nil, err
	}
	return NewPrime256v1Identity(privateKey), nil
}

// PublicKey returns the public key of the identity.
func (id Prime256v1Identity) PublicKey() []byte {
	der, _ := derEncodePrime256v1PublicKey(id.publicKey)
	return der
}

// Sender returns the principal of the identity.
func (id Prime256v1Identity) Sender() principal.Principal {
	return principal.NewSelfAuthenticating(id.PublicKey())
}

// Sign signs the given message. The signature is normalized to low-S form
// (s <= n/2) to match the reference ic-agent and avoid malleability. Unlike
// secp256k1, mainnet does accept high-S P-256, so this is conformance, not a
// hard requirement.
func (id Prime256v1Identity) Sign(msg []byte) ([]byte, error) {
	hashData := sha256.Sum256(msg)
	sigR, sigS, err := ecdsa.Sign(rand.Reader, id.privateKey, hashData[:])
	if err != nil {
		return nil, err
	}
	if sigS.Cmp(p256HalfOrder) == 1 {
		sigS.Sub(p256Order, sigS)
	}
	var buffer [64]byte
	r := sigR.Bytes()
	s := sigS.Bytes()
	copy(buffer[(32-len(r)):], r)
	copy(buffer[(64-len(s)):], s)
	return buffer[:], nil
}

// ToPEM returns the PEM encoding of the private key.
func (id Prime256v1Identity) ToPEM() ([]byte, error) {
	data, err := x509.MarshalECPrivateKey(id.privateKey)
	if err != nil {
		return nil, err
	}
	return pem.EncodeToMemory(&pem.Block{
		Type:  "EC PRIVATE KEY",
		Bytes: data,
	}), nil
}

// Verify verifies the signature of the given message.
func (id Prime256v1Identity) Verify(msg, sig []byte) bool {
	r := new(big.Int).SetBytes(sig[:32])
	s := new(big.Int).SetBytes(sig[32:])
	hashData := sha256.Sum256(msg)
	return ecdsa.Verify(id.publicKey, hashData[:], r, s)
}
