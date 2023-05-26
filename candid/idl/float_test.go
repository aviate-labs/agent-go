package idl_test

import (
	"github.com/aviate-labs/agent-go/candid/idl"
	"testing"
)

func ExampleFloat32Type() {
	test([]idl.Type{idl.Float32Type()}, []any{float32(-0.5)})
	test([]idl.Type{idl.Float32Type()}, []any{float32(0)})
	test([]idl.Type{idl.Float32Type()}, []any{float32(0.5)})
	test([]idl.Type{idl.Float32Type()}, []any{float32(3)})
	// Output:
	// 4449444c000173000000bf
	// 4449444c00017300000000
	// 4449444c0001730000003f
	// 4449444c00017300004040
}

func ExampleFloat64Type() {
	test([]idl.Type{idl.Float64Type()}, []any{-0.5})
	test([]idl.Type{idl.Float64Type()}, []any{float32(0)})
	test([]idl.Type{idl.Float64Type()}, []any{0.5})
	test([]idl.Type{idl.Float64Type()}, []any{float64(3)})
	// Output:
	// 4449444c000172000000000000e0bf
	// 4449444c0001720000000000000000
	// 4449444c000172000000000000e03f
	// 4449444c0001720000000000000840
}

func TestFloatType_UnmarshalGo(t *testing.T) {
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

	t.Run("float64", func(t *testing.T) {
		nt := idl.Float64Type()

		var n float64
		for i, v := range []any{
			float64(0), float32(1),
		} {
			if err := nt.UnmarshalGo(v, &n); err != nil {
				t.Fatal(err)
			}
			if n != float64(i) {
				t.Error(n)
			}
		}

		var a any
		expectErr(t, nt.UnmarshalGo(0, &a))
	})
	t.Run("float32", func(t *testing.T) {
		nt := idl.Float32Type()

		var n float32
		for i, v := range []any{
			float32(0),
		} {
			if err := nt.UnmarshalGo(v, &n); err != nil {
				t.Fatal(err)
			}
			if n != float32(i) {
				t.Error(n)
			}
		}

		for _, v := range []any{
			float64(0),
		} {
			expectErr(t, nt.UnmarshalGo(v, &n))
		}

		var a any
		expectErr(t, nt.UnmarshalGo(0, &a))
	})
}
