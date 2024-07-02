package ii

import (
	"bytes"
	"encoding/hex"
	"github.com/aviate-labs/agent-go/principal"
	"testing"
)

var (
	testCanisterID                 = principal.MustDecode("rwlgt-iiaaa-aaaaa-aaaaa-cai")
	testSeed                       = []byte{42, 72, 44}
	testCanisterSigPublicKeyDER, _ = hex.DecodeString("301f300c060a2b0601040183b8430102030f000a000000000000000001012a482c")
)

func TestCanisterSigPublicKeyFromDER(t *testing.T) {
	cspk, err := CanisterSigPublicKeyFromDER(testCanisterSigPublicKeyDER)
	if err != nil {
		t.Fatal(err)
	}

	if !bytes.Equal(cspk.CanisterID.Raw, testCanisterID.Raw) {
		t.Fatalf("expected %x, got %x", testCanisterID.Raw, cspk.CanisterID.Raw)
	}
	if !bytes.Equal(cspk.Seed, testSeed) {
		t.Fatalf("expected %x, got %x", testSeed, cspk.Seed)
	}
}

func TestCanisterSigPublicKey_DER(t *testing.T) {
	cspk := CanisterSigPublicKey{
		CanisterID: testCanisterID,
		Seed:       testSeed,
	}

	der := cspk.DER()
	if !bytes.Equal(der, testCanisterSigPublicKeyDER) {
		t.Fatalf("expected %x, got %x", testCanisterSigPublicKeyDER, der)
	}
}
