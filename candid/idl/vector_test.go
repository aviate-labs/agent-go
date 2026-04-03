package idl_test

import (
	"errors"
	"testing"

	"github.com/niccolofant/agent-go/candid/idl"
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
			var unmarshalGoError *idl.UnmarshalGoError
			ok := errors.As(err, &unmarshalGoError)
			if !ok {
				t.Fatal("expected UnmarshalGoError")
			}
		}
	}
	t.Run("slice", func(t *testing.T) {
		var nv = []idl.Null{{}}
		if err := idl.UnmarshalGo(idl.VectorType{
			Type: idl.NullType{},
		}, nil, &nv); err != nil {
			t.Fatal(err)
		}
		if len(nv) != 0 {
			t.Error(nv)
		}

		if err := idl.UnmarshalGo(idl.VectorType{
			Type: idl.NullType{},
		}, []any{idl.Null{}, nil}, &nv); err != nil {
			t.Fatal(err)
		}
		if len(nv) != 2 {
			t.Error(nv)
		}

		if err := idl.UnmarshalGo(idl.VectorType{
			Type: idl.NullType{},
		}, [1]idl.Null{{}}, &nv); err != nil {
			t.Fatal(err)
		}
		if len(nv) != 1 {
			t.Error(nv)
		}

		var a any
		expectErr(t, idl.UnmarshalGo(idl.VectorType{
			Type: idl.NullType{},
		}, true, &a))
	})
	t.Run("array", func(t *testing.T) {
		var nv = [1]idl.Int{}
		if err := idl.UnmarshalGo(idl.VectorType{
			Type: idl.IntType{},
		}, nil, &nv); err != nil {
			t.Fatal(err)
		}

		if err := idl.UnmarshalGo(idl.VectorType{
			Type: idl.IntType{},
		}, []any{0}, &nv); err != nil {
			t.Fatal(err)
		}

		if err := idl.UnmarshalGo(idl.VectorType{
			Type: idl.IntType{},
		}, [1]idl.Int{idl.NewInt(0)}, &nv); err != nil {
			t.Fatal(err)
		}

		expectErr(t, idl.UnmarshalGo(idl.VectorType{
			Type: idl.IntType{},
		}, []any{}, &nv))

		expectErr(t, idl.UnmarshalGo(idl.VectorType{
			Type: idl.IntType{},
		}, [2]any{}, &nv))
	})
}

func TestVectorType_empty(t *testing.T) {
	typ := idl.VectorType{Type: idl.Nat8Type()}

	var x []byte
	t.Run("non-nil", func(t *testing.T) {
		if err := idl.UnmarshalGo(typ, []byte{}, &x); err != nil {
			t.Fatal(err)
		}
		if x == nil {
			t.Error("expected non-nil")
		}
	})
	t.Run("nil", func(t *testing.T) {
		if err := idl.UnmarshalGo(typ, nil, &x); err != nil {
			t.Fatal(err)
		}
		if x != nil {
			t.Error("expected nil")
		}
	})
}
