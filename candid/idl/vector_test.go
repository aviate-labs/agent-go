package idl_test

import (
	"github.com/aviate-labs/agent-go/candid/idl"
	"testing"
)

func ExampleVectorType() {
	test([]idl.Type{idl.NewVectorType(new(idl.IntType))}, []any{
		[]any{idl.NewInt(0), idl.NewInt(1), idl.NewInt(2), idl.NewInt(3)},
	})
	// Output:
	// 4449444c016d7c01000400010203
}

func TestVectorType_UnmarshalGo(t *testing.T) {
	expectErr := func(t *testing.T, err error) {
		if err == nil {
			t.Fatal("expected error")
		} else {
			_, ok := err.(*idl.UnmarshalGoError)
			if !ok {
				t.Fatal("expected UnmarshalGoError")
			}
		}
	}
	t.Run("slice", func(t *testing.T) {
		var nv = []idl.Null{{}}
		if err := (idl.VectorType{
			Type: idl.NullType{},
		}).UnmarshalGo(nil, &nv); err != nil {
			t.Fatal(err)
		}
		if len(nv) != 0 {
			t.Error(nv)
		}

		if err := (idl.VectorType{
			Type: idl.NullType{},
		}).UnmarshalGo([]any{idl.Null{}, nil}, &nv); err != nil {
			t.Fatal(err)
		}
		if len(nv) != 2 {
			t.Error(nv)
		}

		if err := (idl.VectorType{
			Type: idl.NullType{},
		}).UnmarshalGo([1]idl.Null{{}}, &nv); err != nil {
			t.Fatal(err)
		}
		if len(nv) != 1 {
			t.Error(nv)
		}

		var a any
		expectErr(t, (idl.VectorType{
			Type: idl.NullType{},
		}).UnmarshalGo(true, &a))
	})
	t.Run("array", func(t *testing.T) {
		var nv = [1]idl.Int{}
		if err := (idl.VectorType{
			Type: idl.IntType{},
		}).UnmarshalGo(nil, &nv); err != nil {
			t.Fatal(err)
		}

		if err := (idl.VectorType{
			Type: idl.IntType{},
		}).UnmarshalGo([]any{0}, &nv); err != nil {
			t.Fatal(err)
		}

		if err := (idl.VectorType{
			Type: idl.IntType{},
		}).UnmarshalGo([1]idl.Int{idl.NewInt(0)}, &nv); err != nil {
			t.Fatal(err)
		}

		expectErr(t, (idl.VectorType{
			Type: idl.IntType{},
		}).UnmarshalGo([]any{}, &nv))

		expectErr(t, (idl.VectorType{
			Type: idl.IntType{},
		}).UnmarshalGo([2]any{}, &nv))
	})
}
