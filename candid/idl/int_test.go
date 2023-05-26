package idl_test

import (
	"math/big"
	"testing"

	"github.com/aviate-labs/agent-go/candid/idl"
)

func ExampleInt() {
	test([]idl.Type{new(idl.IntType)}, []any{idl.NewInt(0)})
	test([]idl.Type{new(idl.IntType)}, []any{idl.NewInt(42)})
	test([]idl.Type{new(idl.IntType)}, []any{idl.NewInt(1234567890)})
	test([]idl.Type{new(idl.IntType)}, []any{idl.NewInt(-1234567890)})
	test([]idl.Type{new(idl.IntType)}, []any{func() idl.Int {
		bi, _ := new(big.Int).SetString("60000000000000000", 10)
		return idl.NewBigInt(bi)
	}()})
	// Output:
	// 4449444c00017c00
	// 4449444c00017c2a
	// 4449444c00017cd285d8cc04
	// 4449444c00017caefaa7b37b
	// 4449444c00017c808098f4e9b5caea00
}

func ExampleInt32Type() {
	test([]idl.Type{idl.Int32Type()}, []any{int32(-1234567890)})
	test([]idl.Type{idl.Int32Type()}, []any{int32(-42)})
	test([]idl.Type{idl.Int32Type()}, []any{int32(42)})
	test([]idl.Type{idl.Int32Type()}, []any{int32(1234567890)})
	// Output:
	// 4449444c0001752efd69b6
	// 4449444c000175d6ffffff
	// 4449444c0001752a000000
	// 4449444c000175d2029649
}

func ExampleInt8Type() {
	test([]idl.Type{idl.Int8Type()}, []any{int16(-129)})
	test([]idl.Type{idl.Int8Type()}, []any{int8(-128)})
	test([]idl.Type{idl.Int8Type()}, []any{int8(-42)})
	test([]idl.Type{idl.Int8Type()}, []any{int8(-1)})
	test([]idl.Type{idl.Int8Type()}, []any{int8(0)})
	test([]idl.Type{idl.Int8Type()}, []any{int8(1)})
	test([]idl.Type{idl.Int8Type()}, []any{int8(42)})
	test([]idl.Type{idl.Int8Type()}, []any{int8(127)})
	test([]idl.Type{idl.Int8Type()}, []any{int16(128)})
	// Output:
	// enc: invalid value: -129
	// 4449444c00017780
	// 4449444c000177d6
	// 4449444c000177ff
	// 4449444c00017700
	// 4449444c00017701
	// 4449444c0001772a
	// 4449444c0001777f
	// enc: invalid value: 128
}

func TestIntType_UnmarshalGo(t *testing.T) {
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

	t.Run("int", func(t *testing.T) {
		var nt idl.IntType

		var n idl.Int
		for i, v := range []any{
			idl.NewInt(0),
			1, int64(2), int32(3), int16(4), int8(5),
		} {
			if err := nt.UnmarshalGo(v, &n); err != nil {
				t.Fatal(err)
			}
			if n.BigInt().Int64() != int64(i) {
				t.Error(n)
			}
		}

		var a any
		expectErr(t, nt.UnmarshalGo(0, &a))
	})
	t.Run("int64", func(t *testing.T) {
		nt := idl.Int64Type()

		var n int64
		for i, v := range []any{
			int64(0), int32(1), int16(2), int8(3),
		} {
			if err := nt.UnmarshalGo(v, &n); err != nil {
				t.Fatal(err)
			}
			if n != int64(i) {
				t.Error(n)
			}
		}

		for _, v := range []any{
			idl.NewInt(0), 0,
		} {
			expectErr(t, nt.UnmarshalGo(v, &n))
		}

		var a any
		expectErr(t, nt.UnmarshalGo(0, &a))
	})
	t.Run("int32", func(t *testing.T) {
		nt := idl.Int32Type()

		var n int32
		for i, v := range []any{
			int32(0), int16(1), int8(2),
		} {
			if err := nt.UnmarshalGo(v, &n); err != nil {
				t.Fatal(err)
			}
			if n != int32(i) {
				t.Error(n)
			}
		}

		for _, v := range []any{
			idl.NewInt(0), 0, int64(0),
		} {
			expectErr(t, nt.UnmarshalGo(v, &n))
		}

		var a any
		expectErr(t, nt.UnmarshalGo(0, &a))
	})
	t.Run("int16", func(t *testing.T) {
		nt := idl.Int16Type()

		var n int16
		for i, v := range []any{
			int16(0), int8(1),
		} {
			if err := nt.UnmarshalGo(v, &n); err != nil {
				t.Fatal(err)
			}
			if n != int16(i) {
				t.Error(n)
			}
		}

		for _, v := range []any{
			idl.NewInt(0), 0, int64(0), int32(0),
		} {
			expectErr(t, nt.UnmarshalGo(v, &n))
		}

		var a any
		expectErr(t, nt.UnmarshalGo(0, &a))
	})
	t.Run("int8", func(t *testing.T) {
		nt := idl.Int8Type()

		var n int8
		for i, v := range []any{
			int8(0),
		} {
			if err := nt.UnmarshalGo(v, &n); err != nil {
				t.Fatal(err)
			}
			if n != int8(i) {
				t.Error(n)
			}
		}

		for _, v := range []any{
			idl.NewInt(0), 0, int64(0), int32(0), int16(0),
		} {
			expectErr(t, nt.UnmarshalGo(v, &n))
		}

		var a any
		expectErr(t, nt.UnmarshalGo(0, &a))
	})
}
