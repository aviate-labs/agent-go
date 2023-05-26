package idl_test

import (
	"github.com/aviate-labs/agent-go/candid/idl"
	"testing"
)

func ExampleText() {
	test([]idl.Type{new(idl.TextType)}, []any{""})
	test([]idl.Type{new(idl.TextType)}, []any{"Motoko"})
	test([]idl.Type{new(idl.TextType)}, []any{"Hi ☃\n"})
	// Output:
	// 4449444c00017100
	// 4449444c000171064d6f746f6b6f
	// 4449444c00017107486920e298830a
}

func TestTextType_UnmarshalGo(t *testing.T) {
	var nt idl.TextType

	var s string
	if err := nt.UnmarshalGo("ok", &s); err != nil {
		t.Fatal(err)
	}
	if s != "ok" {
		t.Error(s)
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
