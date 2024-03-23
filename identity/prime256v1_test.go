package identity

import (
	"bytes"
	"encoding/hex"
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
	if !bytes.Equal(id.privateKey.D.Bytes(), id_.privateKey.D.Bytes()) {
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
	if !id.Verify(data, id.Sign(data)) {
		t.Error()
	}
}
