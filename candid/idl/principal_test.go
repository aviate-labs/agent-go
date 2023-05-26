package idl_test

import (
	"bytes"
	"github.com/aviate-labs/agent-go/candid/idl"
	"github.com/aviate-labs/agent-go/principal"
	"testing"
)

func ExamplePrincipal() {
	p, _ := principal.Decode("aaaaa-aa")
	test([]idl.Type{idl.NewOptionalType(new(idl.PrincipalType))}, []any{p})
	// Output:
	// 4449444c016e680100010100
}

func TestPrincipalType_UnmarshalGo(t *testing.T) {
	var nt idl.PrincipalType

	var p principal.Principal
	if err := nt.UnmarshalGo(principal.AnonymousID, &p); err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(p.Raw, principal.AnonymousID.Raw) {
		t.Error(p)
	}
	var empty []byte
	if err := nt.UnmarshalGo(empty, &p); err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(p.Raw, empty) {
		t.Error(p)
	}

	var a any
	if err := nt.UnmarshalGo(true, &a); err == nil {
		t.Fatal("expected error")
	} else {
		if _, ok := err.(*idl.UnmarshalGoError); !ok {
			t.Fatal("expected UnmarshalGoError")
		}
	}
}
