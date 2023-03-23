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

func ExampleRecordType_nested() {
	recordType := idl.NewRecordType(map[string]idl.Type{
		"foo": idl.Int32Type(),
		"bar": new(idl.BoolType),
	})
	recordValue := map[string]any{
		"foo": int32(42),
		"bar": true,
	}
	test_([]idl.Type{idl.NewRecordType(map[string]idl.Type{
		"foo": idl.Int32Type(),
		"bar": recordType,
		"baz": recordType,
		"bib": recordType,
	})}, []any{
		map[string]any{
			"foo": int32(42),
			"bar": recordValue,
			"baz": recordValue,
			"bib": recordValue,
		},
	})
	// Output:
	// 4449444c026c02d3e3aa027e868eb702756c04d3e3aa0200dbe3aa0200bbf1aa0200868eb702750101012a000000012a000000012a0000002a000000
}
