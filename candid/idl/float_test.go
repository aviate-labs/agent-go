package idl_test

import (
	"github.com/aviate-labs/agent-go/candid/idl"
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
