package idl_test

import (
	"github.com/aviate-labs/agent-go/candid/idl"
	"testing"
)

func ExampleNull() {
	test([]idl.Type{new(idl.NullType)}, []any{nil})
	// Output:
	// 4449444c00017f
}

func TestNullType_UnmarshalGo(t *testing.T) {
	var nt idl.NullType

	var null idl.Null
	if err := nt.UnmarshalGo(nil, &null); err != nil {
		t.Fatal(err)
	}
	if err := nt.UnmarshalGo(idl.Null{}, &null); err != nil {
		t.Fatal(err)
	}
}
