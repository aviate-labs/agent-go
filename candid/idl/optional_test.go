package idl_test

import (
	"github.com/aviate-labs/agent-go/candid/idl"
	"testing"
)

func ExampleOpt() {
	var optNat = idl.NewOptionalType(new(idl.NatType))
	test([]idl.Type{optNat}, []any{nil})
	test([]idl.Type{optNat}, []any{idl.NewNat(uint(1))})
	// Output:
	// 4449444c016e7d010000
	// 4449444c016e7d01000101
}

func TestOptionalType_UnmarshalGo(t *testing.T) {
	var null *idl.Null
	if err := (idl.OptionalType{
		Type: new(idl.NullType),
	}).UnmarshalGo(nil, &null); err != nil {
		t.Fatal(err)
	}

	var nat *idl.Nat
	for i := 0; i < 3; i++ {
		if err := (idl.OptionalType{
			Type: new(idl.NatType),
		}).UnmarshalGo(uint(1), &nat); err != nil {
			t.Fatal(err)
		}
		if nat == nil {
			t.Fatal("expected non-nil")
		}
		if (*nat).BigInt().Int64() != int64(1) {
			t.Fatal(nat)
		}
	}

	var a any
	if err := (idl.OptionalType{
		Type: new(idl.NullType),
	}).UnmarshalGo("", &a); err == nil {
		t.Fatal("expected error")
	} else {
		if _, ok := err.(*idl.UnmarshalGoError); !ok {
			t.Fatal("expected UnmarshalGoError")
		}
	}
}
