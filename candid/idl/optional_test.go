package idl_test

import (
	"errors"
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

func ExampleOpt_blob() {
	var optNatArray = idl.NewOptionalType(idl.VectorType{Type: idl.Nat8Type()})
	test([]idl.Type{optNatArray}, []any{nil})
	test([]idl.Type{optNatArray}, []any{[]byte{0x00}})
	// Output:
	// 4449444c026d7b6e00010100
	// 4449444c026d7b6e000101010100
}

func TestOptionalType_UnmarshalGo(t *testing.T) {
	if err := (idl.OptionalType{
		Type: new(idl.NullType),
	}).UnmarshalGo(nil, new(idl.Null)); err != nil {
		t.Fatal(err)
	}

	var nat *idl.Nat
	for range 3 {
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
		var unmarshalGoError *idl.UnmarshalGoError
		if !errors.As(err, &unmarshalGoError) {
			t.Fatal("expected UnmarshalGoError")
		}
	}

	t.Run("Blob", func(t *testing.T) {
		var bs *[]byte
		if err := (idl.OptionalType{
			Type: idl.NewVectorType(idl.Nat8Type()),
		}).UnmarshalGo([]any{byte(0x00)}, &bs); err != nil {
			t.Error(err)
		}
	})
}
