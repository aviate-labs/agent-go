package idl_test

import (
	"github.com/aviate-labs/agent-go/candid/idl"
)

func ExampleVectorType() {
	test([]idl.Type{idl.NewVectorType(new(idl.IntType))}, []any{
		[]any{idl.NewInt(0), idl.NewInt(1), idl.NewInt(2), idl.NewInt(3)},
	})
	// Output:
	// 4449444c016d7c01000400010203
}
