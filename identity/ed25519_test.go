package identity

import (
	"bytes"
	"testing"
)

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
