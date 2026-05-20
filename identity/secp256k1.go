package identity

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/asn1"
	"encoding/pem"
	"errors"
	"fmt"
	"math/big"
	"slices"

	"github.com/consensys/gnark-crypto/ecc/secp256k1"
	"github.com/consensys/gnark-crypto/ecc/secp256k1/ecdsa"
	"github.com/consensys/gnark-crypto/ecc/secp256k1/fp"
	"github.com/consensys/gnark-crypto/ecc/secp256k1/fr"
	"github.com/niccolofant/agent-go/principal"
)

// Sizes for secp256k1, sourced from gnark-crypto so they stay in lockstep
// with the backing implementation. fp = base field (X, Y); fr = scalar field
// (private key, r, s of a signature).
const (
	coordLen             = fp.Bytes       // one X or Y coordinate.
	scalarLen            = fr.Bytes       // scalar (private key, r, s).
	uncompressedPointLen = 1 + 2*coordLen // 0x04 || X || Y.
)

// The scalar-field modulus equals the secp256k1 curve order.
var (
	secp256k1Order     = fr.Modulus()
	secp256k1HalfOrder = new(big.Int).Rsh(secp256k1Order, 1)
)

var ecPublicKeyOID = asn1.ObjectIdentifier{1, 2, 840, 10045, 2, 1}

var secp256k1OID = asn1.ObjectIdentifier{1, 3, 132, 0, 10}

func derEncodeSecp256k1PublicKey(key *ecdsa.PublicKey) ([]byte, error) {
	return asn1.Marshal(ecPublicKey{
		Metadata: []asn1.ObjectIdentifier{
			ecPublicKeyOID,
			secp256k1OID,
		},
		PublicKey: asn1.BitString{
			Bytes: uncompressedPublicKey(key),
		},
	})
}

func isSecp256k1(actual asn1.ObjectIdentifier) bool {
	return slices.Equal(actual, secp256k1OID)
}

// newSecp256k1PrivateKeyFromASN1 assembles an ecdsa.PrivateKey from an
// SEC1-format ECPrivateKey. gnark's PrivateKey.SetBytes expects X||Y||scalar.
// The SEC1 publicKey field is OPTIONAL (RFC 5915 sec. 3); if absent, we
// derive the point from the scalar via [d]G.
func newSecp256k1PrivateKeyFromASN1(raw ecPrivateKey) (*ecdsa.PrivateKey, error) {
	if len(raw.PrivateKey) == 0 || len(raw.PrivateKey) > scalarLen {
		return nil, fmt.Errorf("scalar length out of range: %d bytes (max %d)", len(raw.PrivateKey), scalarLen)
	}
	var buf [2*coordLen + scalarLen]byte
	switch {
	case len(raw.PublicKey.Bytes) == 0:
		d := new(big.Int).SetBytes(raw.PrivateKey)
		var p secp256k1.G1Affine
		p.ScalarMultiplicationBase(d)
		xy := p.RawBytes() // X || Y
		copy(buf[:2*coordLen], xy[:])
	case len(raw.PublicKey.Bytes) == uncompressedPointLen && raw.PublicKey.Bytes[0] == 0x04:
		copy(buf[:2*coordLen], raw.PublicKey.Bytes[1:])
	default:
		return nil, fmt.Errorf("expected uncompressed public key (%d bytes, leading 0x04)", uncompressedPointLen)
	}
	// Left-pad the scalar; SEC1 may omit leading zero bytes.
	copy(buf[len(buf)-len(raw.PrivateKey):], raw.PrivateKey)
	var pk ecdsa.PrivateKey
	if _, err := pk.SetBytes(buf[:]); err != nil {
		return nil, err
	}
	return &pk, nil
}

// uncompressedPublicKey returns the SEC1 uncompressed public key encoding,
// 0x04 || X || Y.
func uncompressedPublicKey(key *ecdsa.PublicKey) []byte {
	pub := key.Bytes() // X || Y, 2*coordLen bytes
	out := make([]byte, uncompressedPointLen)
	out[0] = 0x04
	copy(out[1:], pub)
	return out
}

// Secp256k1Identity is an identity based on a secp256k1 key pair.
type Secp256k1Identity struct {
	privateKey *ecdsa.PrivateKey
	publicKey  *ecdsa.PublicKey
	derPubKey  []byte
}

// NewRandomSecp256k1Identity creates a new identity with a random key pair.
func NewRandomSecp256k1Identity() (*Secp256k1Identity, error) {
	privateKey, err := ecdsa.GenerateKey(rand.Reader)
	if err != nil {
		return nil, err
	}
	return NewSecp256k1Identity(privateKey)
}

// NewSecp256k1Identity creates a new identity based on the given key pair.
func NewSecp256k1Identity(privateKey *ecdsa.PrivateKey) (*Secp256k1Identity, error) {
	der, err := derEncodeSecp256k1PublicKey(&privateKey.PublicKey)
	if err != nil {
		return nil, err
	}
	return &Secp256k1Identity{
		privateKey: privateKey,
		publicKey:  &privateKey.PublicKey,
		derPubKey:  der,
	}, nil
}

// NewSecp256k1IdentityFromPEM creates a new identity from the given PEM file.
func NewSecp256k1IdentityFromPEM(data []byte) (*Secp256k1Identity, error) {
	blockParams, remainder := pem.Decode(data)
	if blockParams == nil || blockParams.Type != "EC PARAMETERS" {
		return nil, fmt.Errorf("invalid pem file")
	}
	block, remainder := pem.Decode(remainder)
	if block == nil || block.Type != "EC PRIVATE KEY" || len(remainder) != 0 {
		return nil, fmt.Errorf("invalid pem file")
	}
	return parseSecp256k1PEMBody(block)
}

// NewSecp256k1IdentityFromPEMWithoutParameters creates a new identity from the given PEM file.
func NewSecp256k1IdentityFromPEMWithoutParameters(data []byte) (*Secp256k1Identity, error) {
	block, remainder := pem.Decode(data)
	if block == nil || block.Type != "EC PRIVATE KEY" || len(remainder) != 0 {
		return nil, fmt.Errorf("invalid pem file")
	}
	return parseSecp256k1PEMBody(block)
}

func parseSecp256k1PEMBody(block *pem.Block) (*Secp256k1Identity, error) {
	var raw ecPrivateKey
	if _, err := asn1.Unmarshal(block.Bytes, &raw); err != nil {
		return nil, err
	}
	if !isSecp256k1(raw.NamedCurveOID) {
		return nil, errors.New("invalid curve type")
	}
	priv, err := newSecp256k1PrivateKeyFromASN1(raw)
	if err != nil {
		return nil, err
	}
	return NewSecp256k1Identity(priv)
}

// PublicKey returns the public key of the identity.
func (id Secp256k1Identity) PublicKey() []byte {
	return id.derPubKey
}

// Sender returns the principal of the identity.
func (id Secp256k1Identity) Sender() principal.Principal {
	return principal.NewSelfAuthenticating(id.PublicKey())
}

// Sign signs the given message. The signature is normalized to low-S form
// (s <= n/2): mainnet rejects high-S secp256k1 signatures with a 400 Invalid
// signature, and Go's signer does not normalize. Verified against mainnet.
func (id Secp256k1Identity) Sign(msg []byte) ([]byte, error) {
	sig, err := id.privateKey.Sign(msg, sha256.New())
	if err != nil {
		return nil, err
	}
	s := new(big.Int).SetBytes(sig[scalarLen:])
	if s.Cmp(secp256k1HalfOrder) == 1 {
		s.Sub(secp256k1Order, s)
		s.FillBytes(sig[scalarLen:])
	}
	return sig, nil
}

// ToPEM returns the PEM encoding of the public key.
func (id Secp256k1Identity) ToPEM() ([]byte, error) {
	der1, err := asn1.Marshal(secp256k1OID)
	if err != nil {
		return nil, err
	}
	// PrivateKey.Bytes() lays out PublicKey.Bytes() || scalar. On this curve
	// PublicKey.Bytes() is the uncompressed X || Y (gnark's godoc here is
	// boilerplate from EdDSA and does not describe the secp256k1 layout), so
	// the scalar starts at offset 2*coordLen.
	privBytes := id.privateKey.Bytes()
	scalar := privBytes[2*coordLen:]

	der2, err := asn1.Marshal(ecPrivateKey{
		Version:       1,
		PrivateKey:    scalar,
		NamedCurveOID: secp256k1OID,
		PublicKey: asn1.BitString{
			Bytes: uncompressedPublicKey(id.publicKey),
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
	ok, err := id.publicKey.Verify(sig, msg, sha256.New())
	if err != nil {
		return false
	}
	return ok
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
