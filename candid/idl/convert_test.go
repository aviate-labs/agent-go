package idl_test

import (
	"fmt"
	"testing"

	"github.com/aviate-labs/agent-go/candid/idl"
	"github.com/aviate-labs/agent-go/principal"
)

func ExampleTypeOf() {
	fmt.Println(idl.TypeOf(0))
	fmt.Println(idl.TypeOf(idl.Optional{
		V: 0, T: idl.Int64Type(),
	}))
	fmt.Println(idl.TypeOf([]any{0}))
	fmt.Println(idl.TypeOf(map[string]any{
		"foo": 0,
	}))
	fmt.Println(idl.TypeOf(idl.Variant{
		Name:  "foo",
		Value: 0,
		Type: idl.NewVariantType(map[string]idl.Type{
			"foo": new(idl.NatType),
		}),
	}))
	fmt.Println(idl.TypeOf(principal.Principal{}))
	// Output:
	// int64 <nil>
	// opt int64 <nil>
	// vec int64 <nil>
	// record {foo:int64} <nil>
	// variant {foo:int64} <nil>
	// principal <nil>
}

func TestTypeOf_customOptionalType(t *testing.T) {
	typ, err := idl.TypeOf(ONat{})
	if err != nil {
		t.Fatal(err)
	}
	if typ.String() != "opt nat" {
		t.Error(typ)
	}
}

func TestTypeOf_nonInterfaceSlice(t *testing.T) {
	typ, err := idl.TypeOf([]idl.Nat{})
	if err != nil {
		t.Fatal(err)
	}
	if typ.String() != "vec nat" {
		t.Error(typ)
	}
}

type ONat struct {
	V *idl.Nat
}

func (n ONat) SubType() idl.Type {
	return new(idl.NatType)
}

func (n ONat) Value() any {
	return n.V
}
