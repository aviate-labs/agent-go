package idl_test

import (
	"github.com/aviate-labs/agent-go/candid/idl"
	"github.com/aviate-labs/agent-go/principal"
)

func ExampleFunctionType() {
	p, _ := principal.Decode("w7x7r-cok77-xa")
	test_(
		[]idl.Type{
			idl.NewFunctionType(
				[]idl.Type{new(idl.TextType)},
				[]idl.Type{new(idl.NatType)},
				nil,
			),
		},
		[]any{
			idl.PrincipalMethod{
				Principal: p,
				Method:    "foo",
			},
		},
	)
	// Output:
	// 4449444c016a0171017d000100010103caffee03666f6f
}
