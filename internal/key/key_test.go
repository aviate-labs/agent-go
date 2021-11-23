package key

import (
	"bytes"
	"testing"

	"github.com/aviate-labs/bip39"
)

func TestKeys(t *testing.T) {
	e, _ := bip39.NewEntropy(128)
	m, _ := bip39.English.NewMnemonic(e)
	n, _ := New(m, "")
	priv, _, err := Keys(n)
	if err != nil {
		t.Fatal()
	}
	p0, _ := n.Serialize()
	if p := priv.Serialize(); !bytes.Equal(p0, p) {
		t.Error(p)
	}
}
