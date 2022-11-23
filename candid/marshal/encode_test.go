package marshal_test

import (
	"fmt"

	"github.com/aviate-labs/agent-go/candid"
	"github.com/aviate-labs/agent-go/candid/idl"
	"github.com/aviate-labs/agent-go/candid/marshal"
	"github.com/aviate-labs/agent-go/principal"
)

func ExampleMarshal_bool() {
	fmt.Println(idl.Encode([]idl.Type{new(idl.BoolType)}, []any{true}))
	fmt.Println(candid.EncodeValue("(true)"))
	fmt.Println(marshal.Marshal([]any{true}))
	// Output:
	// [68 73 68 76 0 1 126 1] <nil>
	// [68 73 68 76 0 1 126 1] <nil>
	// [68 73 68 76 0 1 126 1] <nil>
}

func ExampleMarshal_empty() {
	fmt.Println(marshal.Marshal([]any{idl.Empty{}}))
	// Output:
	// [68 73 68 76 0 1 111] <nil>
}

func ExampleMarshal_nat() {
	fmt.Println(idl.Encode([]idl.Type{new(idl.NatType)}, []any{idl.NewNat(uint(5))}))
	fmt.Println(candid.EncodeValue("(5 : nat)"))
	fmt.Println(marshal.Marshal([]any{idl.NewNat(uint(5))}))
	// Output:
	// [68 73 68 76 0 1 125 5] <nil>
	// [68 73 68 76 0 1 125 5] <nil>
	// [68 73 68 76 0 1 125 5] <nil>
}

func ExampleMarshal_null() {
	fmt.Println(marshal.Marshal([]any{idl.Null{}}))
	// Output:
	// [68 73 68 76 0 1 127] <nil>
}

func ExampleMarshal_principal() {
	p, _ := principal.Decode("aaaaa-aa")
	fmt.Println(marshal.Marshal([]any{&p}))
	fmt.Println(marshal.Marshal([]any{p}))
	// Output:
	// [68 73 68 76 0 1 104 1 0] <nil>
	// [68 73 68 76 0 1 104 1 0] <nil>
}

func ExampleMarshal_record() {
	fmt.Println(idl.Encode([]idl.Type{idl.NewRecordType(map[string]idl.Type{
		"foo": new(idl.TextType),
		"bar": new(idl.IntType),
	})}, []any{
		map[string]any{
			"foo": "baz",
			"bar": idl.NewInt(42),
		},
	}))
	fmt.Println(marshal.Marshal([]any{map[string]any{
		"foo": "baz",
		"bar": idl.NewInt(42),
	}}))
	// Output:
	// [68 73 68 76 1 108 2 211 227 170 2 124 134 142 183 2 113 1 0 42 3 98 97 122] <nil>
	// [68 73 68 76 1 108 2 211 227 170 2 124 134 142 183 2 113 1 0 42 3 98 97 122] <nil>
}

func ExampleMarshal_reserved() {
	fmt.Println(marshal.Marshal([]any{idl.Reserved{}}))
	// Output:
	// [68 73 68 76 0 1 112] <nil>
}
