package principal_test

import (
	"bytes"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"testing"

	"github.com/aviate-labs/agent-go/principal"
)

var LEDGER_PRINCIPAL = principal.MustDecode("ryjl3-tyaaa-aaaaa-aaaba-cai")

func ExampleDecode() {
	p, _ := principal.Decode("em77e-bvlzu-aq")
	fmt.Printf("%x", p.Raw)
	// Output:
	// abcd01
}

func ExamplePrincipal() {
	raw, _ := hex.DecodeString("abcd01")
	p := principal.Principal{Raw: raw}
	fmt.Println(p.Encode())
	// Output:
	// em77e-bvlzu-aq
}

func TestPrincipal(t *testing.T) {
	if !LEDGER_PRINCIPAL.IsOpaque() {
		t.Fatal("expected opaque principal")
	}
	if !principal.MustDecode("g27xm-fnyhk-uu73a-njpqd-hec7y-syhwe-bd45b-qm6yc-xikg5-cylqt-iae").IsSelfAuthenticating() {
		t.Fatal("expected derived principal")
	}
	if !principal.AnonymousID.IsAnonymous() {
		t.Fatal("expected anonymous principal")
	}
	if !(principal.Principal{Raw: append([]byte("random"), 0x7f)}).IsReserved() {
		t.Fatal("expected reserved principal")
	}
}

func TestPrincipal_MarshalJSON(t *testing.T) {
	original := principal.MustDecode("em77e-bvlzu-aq")
	raw, err := original.MarshalJSON()
	if err != nil {
		t.Error(err)
	}
	var decoded principal.Principal
	if err := json.Unmarshal(raw, &decoded); err != nil {
		t.Error(err)
	}
	if !bytes.Equal(original.Raw, decoded.Raw) {
		t.Errorf("expected %v, got %v", original, decoded)
	}
}
