package idl_test

import (
	"testing"

	"github.com/niccolofant/agent-go/candid/idl"
)

func ExampleNull() {
	test([]idl.Type{new(idl.NullType)}, []any{nil})
	// Output:
	// 4449444c00017f
}

func TestNullType_UnmarshalGo(t *testing.T) {
	var nt idl.NullType

	var null idl.Null
	if err := idl.UnmarshalGo(nt, nil, &null); err != nil {
		t.Fatal(err)
	}
	if err := idl.UnmarshalGo(nt, idl.Null{}, &null); err != nil {
		t.Fatal(err)
	}
}
