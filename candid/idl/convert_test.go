package idl_test

import (
	"fmt"
	"testing"

	"github.com/aviate-labs/agent-go/candid/idl"
	"github.com/aviate-labs/agent-go/principal"
)

func ExampleTypeOf() {
	i := 0
	fmt.Println(idl.TypeOf(i))
	fmt.Println(idl.TypeOf(&i))
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

func TestTypeOf_nil(t *testing.T) {
	if typ, err := idl.TypeOf(nil); err != nil {
		t.Fatal(err)
	} else {
		if typ.String() != "null" {
			t.Error(typ)
		}
	}

	var x *int
	typ, err := idl.TypeOf(x)
	if err != nil {
		t.Fatal(err)
	}
	if typ.String() != "opt int64" {
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
