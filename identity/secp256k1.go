package identity

import (
	"crypto/sha256"
	"encoding/asn1"
	"encoding/pem"
	"errors"
	"fmt"
	"github.com/aviate-labs/agent-go/principal"
	"github.com/aviate-labs/secp256k1"
	"math/big"
	"slices"
)

var ecPublicKeyOID = asn1.ObjectIdentifier{1, 2, 840, 10045, 2, 1}

var secp256k1OID = asn1.ObjectIdentifier{1, 3, 132, 0, 10}

func derEncodeSecp256k1PublicKey(key *secp256k1.PublicKey) ([]byte, error) {
	point := key.ToECDSA()
	return asn1.Marshal(ecPublicKey{
		Metadata: []asn1.ObjectIdentifier{
			ecPublicKeyOID,
			secp256k1OID,
		},
		PublicKey: asn1.BitString{
			Bytes: marshal(secp256k1.S256(), point.X, point.Y),
		},
	})
}

func isSecp256k1(actual asn1.ObjectIdentifier) bool {
	return slices.Equal(actual, secp256k1OID)
}

// Secp256k1Identity is an identity based on a secp256k1 key pair.
type Secp256k1Identity struct {
	privateKey *secp256k1.PrivateKey
	publicKey  *secp256k1.PublicKey
}

// NewRandomSecp256k1Identity creates a new identity with a random key pair.
func NewRandomSecp256k1Identity() (*Secp256k1Identity, error) {
	privateKey, err := secp256k1.NewPrivateKey(secp256k1.S256())
	if err != nil {
		return nil, err
	}
	return NewSecp256k1Identity(privateKey)
}

// NewSecp256k1Identity creates a new identity based on the given key pair.
func NewSecp256k1Identity(privateKey *secp256k1.PrivateKey) (*Secp256k1Identity, error) {
	return &Secp256k1Identity{
		privateKey: privateKey,
		publicKey:  privateKey.PubKey(),
	}, nil
}

// NewSecp256k1IdentityFromPEM creates a new identity from the given PEM file.
func NewSecp256k1IdentityFromPEM(data []byte) (*Secp256k1Identity, error) {
	blockParams, remainder := pem.Decode(data)
	if blockParams == nil || blockParams.Type != "EC PARAMETERS" {
		return nil, fmt.Errorf("invalid pem file")
	}
	block, remainder := pem.Decode(remainder)
	if block == nil || blockParams.Type != "EC PARAMETERS" || len(remainder) != 0 {
		return nil, fmt.Errorf("invalid pem file")
	}
	var ecPrivateKey ecPrivateKey
	if _, err := asn1.Unmarshal(block.Bytes, &ecPrivateKey); err != nil {
		return nil, err
	}
	if !isSecp256k1(ecPrivateKey.NamedCurveOID) {
		return nil, errors.New("invalid curve type")
	}
	privateKey, _ := secp256k1.PrivKeyFromBytes(secp256k1.S256(), ecPrivateKey.PrivateKey)
	return NewSecp256k1Identity(privateKey)
}

// NewSecp256k1IdentityFromPEMWithoutParameters creates a new identity from the given PEM file.
func NewSecp256k1IdentityFromPEMWithoutParameters(data []byte) (*Secp256k1Identity, error) {
	block, remainder := pem.Decode(data)
	if block == nil || block.Type != "EC PRIVATE KEY" || len(remainder) != 0 {
		return nil, fmt.Errorf("invalid pem file")
	}
	var ecPrivateKey ecPrivateKey
	if _, err := asn1.Unmarshal(block.Bytes, &ecPrivateKey); err != nil {
		return nil, err
	}
	if !isSecp256k1(ecPrivateKey.NamedCurveOID) {
		return nil, errors.New("invalid curve type")
	}
	privateKey, _ := secp256k1.PrivKeyFromBytes(secp256k1.S256(), ecPrivateKey.PrivateKey)
	return NewSecp256k1Identity(privateKey)
}

// PublicKey returns the public key of the identity.
func (id Secp256k1Identity) PublicKey() []byte {
	der, _ := derEncodeSecp256k1PublicKey(id.publicKey)
	return der
}

// Sender returns the principal of the identity.
func (id Secp256k1Identity) Sender() principal.Principal {
	return principal.NewSelfAuthenticating(id.PublicKey())
}

// Sign signs the given message.
func (id Secp256k1Identity) Sign(msg []byte) []byte {
	hashData := sha256.Sum256(msg)
	sig, _ := id.privateKey.Sign(hashData[:])
	var buffer [64]byte
	r := sig.R.Bytes()
	s := sig.S.Bytes()
	copy(buffer[(32-len(r)):], r)
	copy(buffer[(64-len(s)):], s)
	return buffer[:]
}

// ToPEM returns the PEM encoding of the public key.
func (id Secp256k1Identity) ToPEM() ([]byte, error) {
	der1, err := asn1.Marshal(secp256k1OID)
	if err != nil {
		return nil, err
	}
	point := id.publicKey.ToECDSA()
	der2, err := asn1.Marshal(ecPrivateKey{
		Version:       1,
		PrivateKey:    id.privateKey.D.Bytes(),
		NamedCurveOID: secp256k1OID,
		PublicKey: asn1.BitString{
			Bytes: marshal(secp256k1.S256(), point.X, point.Y),
		},
	})
	if err != nil {
		return nil, err
	}
	return append(
		pem.EncodeToMemory(&pem.Block{
			Type:  "EC PARAMETERS",
			Bytes: der1,
		}),
		pem.EncodeToMemory(&pem.Block{
			Type:  "EC PRIVATE KEY",
			Bytes: der2,
		})...,
	), nil
}

// Verify verifies the signature of the given message.
func (id Secp256k1Identity) Verify(msg, sig []byte) bool {
	signature := secp256k1.Signature{
		R: new(big.Int).SetBytes(sig[:32]),
		S: new(big.Int).SetBytes(sig[32:]),
	}
	hashData := sha256.Sum256(msg)
	return signature.Verify(hashData[:], id.publicKey)
}

type ecPrivateKey struct {
	Version       int
	PrivateKey    []byte
	NamedCurveOID asn1.ObjectIdentifier `asn1:"optional,explicit,tag:0"`
	PublicKey     asn1.BitString        `asn1:"optional,explicit,tag:1"`
}

type ecPublicKey struct {
	Metadata  []asn1.ObjectIdentifier
	PublicKey asn1.BitString
}
