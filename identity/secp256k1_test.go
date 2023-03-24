package identity

import (
	"bytes"
	"testing"
)

func TestNewSecp256k1Identity(t *testing.T) {
	id, _ := NewRandomSecp256k1Identity()
	data, err := id.ToPEM()
	if err != nil {
		t.Fatal(err)
	}
	id_, err := NewSecp256k1IdentityFromPEM(data)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(id.privateKey.Serialize(), id_.privateKey.Serialize()) {
		t.Error()
	}
	if !bytes.Equal(id.PublicKey(), id_.PublicKey()) {
		t.Error()
	}
}
