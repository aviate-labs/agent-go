package idl_test

import (
	"github.com/aviate-labs/agent-go/candid/idl"
)

func ExampleBool() {
	test([]idl.Type{new(idl.BoolType)}, []any{true})
	test([]idl.Type{new(idl.BoolType)}, []any{false})
	test([]idl.Type{new(idl.BoolType)}, []any{0})
	test([]idl.Type{new(idl.BoolType)}, []any{"false"})
	// Output:
	// 4449444c00017e01
	// 4449444c00017e00
	// enc: invalid argument: 0
	// enc: invalid argument: false
}
