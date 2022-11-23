package idl_test

import (
	"github.com/aviate-labs/agent-go/candid/idl"
)

func ExampleRecordType() {
	test([]idl.Type{idl.NewRecordType(nil)}, []any{nil})
	test_([]idl.Type{idl.NewRecordType(map[string]idl.Type{
		"foo": new(idl.TextType),
		"bar": new(idl.IntType),
	})}, []any{
		map[string]any{
			"foo": "💩",
			"bar": idl.NewInt(42),
			"baz": idl.NewInt(0),
		},
	})
	// Output:
	// 4449444c016c000100
	// 4449444c016c02d3e3aa027c868eb7027101002a04f09f92a9
}
