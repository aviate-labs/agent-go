package idl_test

import (
	"github.com/aviate-labs/agent-go/candid/idl"
)

func ExampleOpt() {
	var optNat *idl.OptionalType = idl.NewOptionalType(new(idl.NatType))
	test([]idl.Type{optNat}, []any{nil})
	test([]idl.Type{optNat}, []any{idl.NewNat(uint(1))})
	// Output:
	// 4449444c016e7d010000
	// 4449444c016e7d01000101
}
