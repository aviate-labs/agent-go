package identity

import (
	"crypto/elliptic"
	"crypto/sha256"
	"encoding/asn1"
	"encoding/pem"
	"errors"
	"fmt"
	"github.com/aviate-labs/agent-go/principal"
	"github.com/aviate-labs/secp256k1"
	"golang.org/x/exp/slices"
)

var secp256k1OID = asn1.ObjectIdentifier{1, 3, 132, 0, 10}

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

func NewRandomSecp256k1Identity() (*Secp256k1Identity, error) {
	privateKey, err := secp256k1.NewPrivateKey(secp256k1.S256())
	if err != nil {
		return nil, err
	}
	return NewSecp256k1Identity(privateKey)
}

func derEncodeSecp256k1PublicKey(key *secp256k1.PublicKey) ([]byte, error) {
	point := key.ToECDSA()
	return asn1.Marshal(ecPublicKey{
		Metadata: []asn1.ObjectIdentifier{
			{1, 2, 840, 10045, 2, 1}, // ec.PublicKey
			secp256k1OID,             // Secp256k1
		},
		PublicKey: asn1.BitString{
			Bytes: elliptic.Marshal(secp256k1.S256(), point.X, point.Y),
		},
	})
}

func isSecp256k1(actual asn1.ObjectIdentifier) bool {
	return slices.Equal(actual, secp256k1OID)
}

type Secp256k1Identity struct {
	privateKey *secp256k1.PrivateKey
	publicKey  *secp256k1.PublicKey
}

func NewSecp256k1IdentityFromPEM(data []byte) (*Secp256k1Identity, error) {
	blockParams, remainder := pem.Decode(data)
	if blockParams.Type != "EC PARAMETERS" {
		return nil, fmt.Errorf("invalid pem parameters")
	}
	block, _ := pem.Decode(remainder)
	if blockParams.Type != "EC PARAMETERS" {
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

func NewSecp256k1Identity(privateKey *secp256k1.PrivateKey) (*Secp256k1Identity, error) {
	return &Secp256k1Identity{
		privateKey: privateKey,
		publicKey:  privateKey.PubKey(),
	}, nil
}

func (id Secp256k1Identity) Sender() principal.Principal {
	return principal.NewSelfAuthenticating(id.PublicKey())
}

func (id Secp256k1Identity) Sign(msg []byte) []byte {
	hash := sha256.New()
	hash.Write(msg)
	hashData := hash.Sum(nil)
	sig, _ := id.privateKey.Sign(hashData)
	var buffer [64]byte
	r := sig.R.Bytes()
	s := sig.S.Bytes()
	copy(buffer[(32-len(r)):], r)
	copy(buffer[(64-len(s)):], s)
	return buffer[:]
}

func (id Secp256k1Identity) PublicKey() []byte {
	der, _ := derEncodeSecp256k1PublicKey(id.publicKey)
	return der
}

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
			Bytes: elliptic.Marshal(secp256k1.S256(), point.X, point.Y),
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
