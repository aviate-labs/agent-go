package identity

import (
	"bytes"
	"crypto/elliptic"
	"encoding/hex"
	"math/big"
	"testing"
)

func TestNewPrime256v1Identity(t *testing.T) {
	id, _ := NewRandomPrime256v1Identity()
	data, err := id.ToPEM()
	if err != nil {
		t.Fatal(err)
	}
	id_, err := NewPrime256v1IdentityFromPEM(data)
	if err != nil {
		t.Fatal(err)
	}
	b1, err := id.privateKey.Bytes()
	if err != nil {
		t.Fatal(err)
	}
	b2, err := id_.privateKey.Bytes()
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(b1, b2) {
		t.Error()
	}
	if !bytes.Equal(id.PublicKey(), id_.PublicKey()) {
		t.Error()
	}
}

func TestNewPrime256v1IdentityFromPEM(t *testing.T) {
	pem := `
-----BEGIN EC PRIVATE KEY-----
MHcCAQEEIL1ybmbwx+uKYsscOZcv71MmKhrNqfPP0ke1unET5AY4oAoGCCqGSM49
AwEHoUQDQgAEUbbZV4NerZTPWfbQ749/GNLu8TaH8BUS/I7/+ipsu+MPywfnBFIZ
Sks4xGbA/ZbazsrMl4v446U5UIVxCGGaKw==
-----END EC PRIVATE KEY-----
`
	der, _ := hex.DecodeString("3059301306072a8648ce3d020106082a8648ce3d0301070342000451b6d957835ead94cf59f6d0ef8f7f18d2eef13687f01512fc8efffa2a6cbbe30fcb07e70452194a4b38c466c0fd96dacecacc978bf8e3a53950857108619a2b")
	id, err := NewPrime256v1IdentityFromPEM([]byte(pem))
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(id.PublicKey(), der) {
		t.Fatal("public key mismatch")
	}
}

func TestPrime256v1Identity_Sign(t *testing.T) {
	id, err := NewRandomPrime256v1Identity()
	if err != nil {
		t.Fatal(err)
	}
	data := []byte("hello")
	sig, err := id.Sign(data)
	if err != nil {
		t.Fatal(err)
	}
	if !id.Verify(data, sig) {
		t.Error()
	}
}

// The IC expects low-S ECDSA signatures (s <= n/2); Go's ecdsa.Sign does not
// normalize, so Sign must. High-S occurs ~50% of the time, so loop.
func TestPrime256v1Identity_Sign_LowS(t *testing.T) {
	id, err := NewRandomPrime256v1Identity()
	if err != nil {
		t.Fatal(err)
	}
	half := new(big.Int).Rsh(elliptic.P256().Params().N, 1)
	for i := range 64 {
		sig, err := id.Sign([]byte{byte(i)})
		if err != nil {
			t.Fatal(err)
		}
		if len(sig) != 64 {
			t.Fatalf("signature length: got %d, want 64", len(sig))
		}
		if s := new(big.Int).SetBytes(sig[32:]); s.Cmp(half) == 1 {
			t.Fatalf("high-S signature at i=%d: s > n/2", i)
		}
	}
}
