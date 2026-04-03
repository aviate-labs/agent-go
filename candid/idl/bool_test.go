package idl_test

import (
	"testing"

	"github.com/niccolofant/agent-go/candid/idl"
)

func ExampleBool() {
	test([]idl.Type{new(idl.BoolType)}, []any{true})
	test([]idl.Type{new(idl.BoolType)}, []any{false})
	test([]idl.Type{new(idl.BoolType)}, []any{0})
	test([]idl.Type{new(idl.BoolType)}, []any{"false"})
	// Output:
	// 4449444c00017e01
	// 4449444c00017e00
	// enc: invalid type 0 (int), expected type bool
	// enc: invalid type false (string), expected type bool
}

func TestBoolType_UnmarshalGo(t *testing.T) {
	var nt idl.BoolType

	var b bool
	if err := idl.UnmarshalGo(nt, true, &b); err != nil {
		t.Fatal(err)
	}
	if !b {
		t.Error(b)
	}

	var a any
	if err := idl.UnmarshalGo(nt, true, &a); err == nil {
		t.Fatal("expected error")
	} else {
		if _, ok := err.(*idl.UnmarshalGoError); !ok {
			t.Fatal("expected UnmarshalGoError")
		}
	}
}
