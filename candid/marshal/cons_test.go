package marshal_test

import (
	"github.com/aviate-labs/agent-go/candid/idl"
	"github.com/aviate-labs/agent-go/candid/marshal"
)

func ExampleDecodeRecord() {
	ctx := marshal.NewContext()

	printDecode(marshal.DecodeRecord(hexToBytesReader("000100"), marshal.ContextToType(ctx, idl.NewRecordType(nil))))
	printDecode(marshal.DecodeRecord(
		hexToBytesReader("2a04f09f92a9"),
		marshal.ContextToType(ctx, idl.NewRecordType(map[string]idl.Type{
			"foo": new(idl.TextType),
			"bar": new(idl.IntType),
		})),
	))
	// Output:
	// map[]
	// map[bar:42 foo:💩]
}
