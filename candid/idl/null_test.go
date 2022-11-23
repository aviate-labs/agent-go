package idl_test

import (
	"github.com/aviate-labs/agent-go/candid/idl"
)

func ExampleNull() {
	test([]idl.Type{new(idl.NullType)}, []any{nil})
	// Output:
	// 4449444c00017f
}
