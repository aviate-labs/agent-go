package idl_test

import (
	"math/big"

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
