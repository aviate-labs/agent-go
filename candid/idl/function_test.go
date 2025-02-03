package idl_test

import (
	"github.com/aviate-labs/agent-go/candid/idl"
	"github.com/aviate-labs/agent-go/principal"
)

func ExampleFunctionType() {
	test_(
		[]idl.Type{
			idl.NewFunctionType(
				[]idl.FunctionParameter{{Type: new(idl.TextType)}},
				[]idl.FunctionParameter{{Type: new(idl.NatType)}},
				nil,
			),
		},
		[]any{
			&idl.PrincipalMethod{
				Principal: principal.MustDecode("w7x7r-cok77-xa"),
				Method:    "foo",
			},
		},
	)
	// Output:
	// 4449444c016a0171017d000100010103caffee03666f6f
}
