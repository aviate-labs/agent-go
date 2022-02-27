package identity

import (
	"bytes"
	"crypto/ed25519"
	"crypto/rand"
	"testing"
)

func TestNewEd25519Identity(t *testing.T) {
	publicKey, privateKey, _ := ed25519.GenerateKey(rand.Reader)
	id := NewEd25519Identity(publicKey, privateKey)
	data, err := id.ToPEM()
	if err != nil {
		t.Fatal(err)
	}
	id_, err := NewEd25519IdentityFromPEM(data)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(id.privateKey, id_.privateKey) {
		t.Error()
	}
	if !bytes.Equal(id.publicKey, id_.publicKey) {
		t.Error()
	}
}
