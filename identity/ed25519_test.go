package identity

import (
	"bytes"
	"testing"
)

func TestEd25519Identity_Sign(t *testing.T) {
	id, err := NewRandomEd25519Identity()
	if err != nil {
		t.Fatal(err)
	}
	data := []byte("hello")
	if !id.Verify(data, id.Sign(data)) {
		t.Error()
	}
}

func TestNewEd25519Identity(t *testing.T) {
	id, _ := NewRandomEd25519Identity()
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
