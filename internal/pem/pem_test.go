package pem

import (
	"testing"

	"github.com/aviate-labs/agent-go/internal/key"
)

func TestRandomKey(t *testing.T) {
	key, err := key.RandomPrivateKey()
	if err != nil {
		t.Fatal(err)
	}
	e, err := Encode(key)
	if err != nil {
		t.Fatal(err)
	}
	d, err := Decode(e)
	if err != nil {
		t.Fatal(err)
	}
	if key.D.Cmp(d.D) != 0 {
		t.Error(key.D)
		t.Error(d.D)
	}
}
