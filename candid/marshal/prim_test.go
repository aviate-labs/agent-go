package marshal_test

import (
	"github.com/aviate-labs/agent-go/candid/idl"
	"github.com/aviate-labs/agent-go/candid/marshal"
)

func ExampleDecodeBool() {
	printDecode(marshal.DecodeBool(hexToBytesReader("01")))
	printDecode(marshal.DecodeBool(hexToBytesReader("00")))
	// Output:
	// true
	// false
}

func ExampleDecodeFloat32() {
	printDecode(marshal.DecodeFloat32(hexToBytesReader("000000bf")))
	printDecode(marshal.DecodeFloat32(hexToBytesReader("00000000")))
	printDecode(marshal.DecodeFloat32(hexToBytesReader("0000003f")))
	printDecode(marshal.DecodeFloat32(hexToBytesReader("00004040")))
	// Output:
	// -0.5
	// 0
	// 0.5
	// 3
}

func ExampleDecodeFloat64() {
	printDecode(marshal.DecodeFloat64(hexToBytesReader("000000000000e0bf")))
	printDecode(marshal.DecodeFloat64(hexToBytesReader("0000000000000000")))
	printDecode(marshal.DecodeFloat64(hexToBytesReader("000000000000e03f")))
	printDecode(marshal.DecodeFloat64(hexToBytesReader("0000000000000840")))
	// Output:
	// -0.5
	// 0
	// 0.5
	// 3
}

func ExampleDecodeInt() {
	printDecode(marshal.DecodeInt(hexToBytesReader("00")))
	printDecode(marshal.DecodeInt(hexToBytesReader("2a")))
	printDecode(marshal.DecodeInt(hexToBytesReader("d285d8cc04")))
	printDecode(marshal.DecodeInt(hexToBytesReader("aefaa7b37b")))
	printDecode(marshal.DecodeInt(hexToBytesReader("808098f4e9b5caea00")))
	// Output:
	// 0
	// 42
	// 1234567890
	// -1234567890
	// 60000000000000000
}

func ExampleDecodeInt32() {
	printDecode(marshal.DecodeInt32(hexToBytesReader("2efd69b6")))
	printDecode(marshal.DecodeInt32(hexToBytesReader("d6ffffff")))
	printDecode(marshal.DecodeInt32(hexToBytesReader("2a000000")))
	printDecode(marshal.DecodeInt32(hexToBytesReader("d2029649")))
	// Output:
	// -1234567890
	// -42
	// 42
	// 1234567890
}

func ExampleDecodeInt8() {
	printDecode(marshal.DecodeInt8(hexToBytesReader("80")))
	printDecode(marshal.DecodeInt8(hexToBytesReader("d6")))
	printDecode(marshal.DecodeInt8(hexToBytesReader("ff")))
	printDecode(marshal.DecodeInt8(hexToBytesReader("00")))
	printDecode(marshal.DecodeInt8(hexToBytesReader("01")))
	printDecode(marshal.DecodeInt8(hexToBytesReader("2a")))
	printDecode(marshal.DecodeInt8(hexToBytesReader("7f")))
	// Output:
	// -128
	// -42
	// -1
	// 0
	// 1
	// 42
	// 127
}

func ExampleDecodeNat() {
	printDecode(marshal.DecodeNat(hexToBytesReader("00")))
	printDecode(marshal.DecodeNat(hexToBytesReader("2a")))
	printDecode(marshal.DecodeNat(hexToBytesReader("d285d8cc04")))
	printDecode(marshal.DecodeNat(hexToBytesReader("808098f4e9b5ca6a")))
	// Output:
	// 0
	// 42
	// 1234567890
	// 60000000000000000
}

func ExampleDecodeNat16() {
	printDecode(marshal.DecodeNat16(hexToBytesReader("0000")))
	printDecode(marshal.DecodeNat16(hexToBytesReader("2a00")))
	printDecode(marshal.DecodeNat16(hexToBytesReader("ffff")))
	// Output:
	// 0
	// 42
	// 65535
}

func ExampleDecodeNat32() {
	printDecode(marshal.DecodeNat32(hexToBytesReader("00000000")))
	printDecode(marshal.DecodeNat32(hexToBytesReader("2a000000")))
	printDecode(marshal.DecodeNat32(hexToBytesReader("ffffffff")))
	// Output:
	// 0
	// 42
	// 4294967295
}

func ExampleDecodeNat64() {
	printDecode(marshal.DecodeNat64(hexToBytesReader("0000000000000000")))
	printDecode(marshal.DecodeNat64(hexToBytesReader("2a00000000000000")))
	printDecode(marshal.DecodeNat64(hexToBytesReader("d202964900000000")))
	// Output:
	// 0
	// 42
	// 1234567890
}

func ExampleDecodeNat8() {
	printDecode(marshal.DecodeNat8(hexToBytesReader("00")))
	printDecode(marshal.DecodeNat8(hexToBytesReader("2a")))
	printDecode(marshal.DecodeNat8(hexToBytesReader("ff")))
	// Output:
	// 0
	// 42
	// 255
}

func ExampleDecodePrincipal() {
	printDecode(marshal.DecodePrincipal(hexToBytesReader("0100")))
	printDecode(marshal.DecodePrincipal(hexToBytesReader("0103caffee")))
	printDecode(marshal.DecodePrincipal(hexToBytesReader("0109efcdab000000000001")))
	// Output:
	// aaaaa-aa
	// w7x7r-cok77-xa
	// 2chl6-4hpzw-vqaaa-aaaaa-c
}

func ExampleDecodeText() {
	printDecode(marshal.DecodeText(hexToBytesReader("00")))
	printDecode(marshal.DecodeText(hexToBytesReader("064d6f746f6b6f")))
	printDecode(marshal.DecodeText(hexToBytesReader("07486920e298830a")))
	// Output:
	// Motoko
	// Hi ☃
}

func ExampleEncodeBool() {
	printEncode(marshal.EncodeBool(true))
	printEncode(marshal.EncodeBool(false))
	// Output:
	// 7e01
	// 7e00
}

func ExampleEncodeFloat32() {
	printEncode(marshal.EncodeFloat32(-0.5))
	printEncode(marshal.EncodeFloat32(0))
	printEncode(marshal.EncodeFloat32(0.5))
	printEncode(marshal.EncodeFloat32(3))
	// Output:
	// 73000000bf
	// 7300000000
	// 730000003f
	// 7300004040
}

func ExampleEncodeFloat64() {
	printEncode(marshal.EncodeFloat64(-0.5))
	printEncode(marshal.EncodeFloat64(0))
	printEncode(marshal.EncodeFloat64(0.5))
	printEncode(marshal.EncodeFloat64(3))
	// Output:
	// 72000000000000e0bf
	// 720000000000000000
	// 72000000000000e03f
	// 720000000000000840
}

func ExampleEncodeInt() {
	printEncode(marshal.EncodeInt(idl.NewInt(0)))
	printEncode(marshal.EncodeInt(idl.NewInt(42)))
	printEncode(marshal.EncodeInt(idl.NewInt(1234567890)))
	printEncode(marshal.EncodeInt(idl.NewInt(-1234567890)))
	printEncode(marshal.EncodeInt(idl.NewIntFromString("60000000000000000")))
	// Output:
	// 7c00
	// 7c2a
	// 7cd285d8cc04
	// 7caefaa7b37b
	// 7c808098f4e9b5caea00
}

func ExampleEncodeInt32() {
	printEncode(marshal.EncodeInt32(-1234567890))
	printEncode(marshal.EncodeInt32(-42))
	printEncode(marshal.EncodeInt32(42))
	printEncode(marshal.EncodeInt32(1234567890))
	// Output:
	// 752efd69b6
	// 75d6ffffff
	// 752a000000
	// 75d2029649
}

func ExampleEncodeInt8() {
	printEncode(marshal.EncodeInt8(-128))
	printEncode(marshal.EncodeInt8(-42))
	printEncode(marshal.EncodeInt8(-1))
	printEncode(marshal.EncodeInt8(0))
	printEncode(marshal.EncodeInt8(1))
	printEncode(marshal.EncodeInt8(42))
	printEncode(marshal.EncodeInt8(127))
	// Output:
	// 7780
	// 77d6
	// 77ff
	// 7700
	// 7701
	// 772a
	// 777f
}

func ExampleEncodeNat() {
	printEncode(marshal.EncodeNat(idl.NewNat[uint](0)))
	printEncode(marshal.EncodeNat(idl.NewNat[uint](42)))
	printEncode(marshal.EncodeNat(idl.NewNat[uint](1234567890)))
	printEncode(marshal.EncodeNat(idl.NewNatFromString("60000000000000000")))
	// Output:
	// 7d00
	// 7d2a
	// 7dd285d8cc04
	// 7d808098f4e9b5ca6a
}

func ExampleEncodeNat16() {
	printEncode(marshal.EncodeNat16(0))
	printEncode(marshal.EncodeNat16(42))
	printEncode(marshal.EncodeNat16(65535))
	// Output:
	// 7a0000
	// 7a2a00
	// 7affff
}

func ExampleEncodeNat32() {
	printEncode(marshal.EncodeNat32(0))
	printEncode(marshal.EncodeNat32(42))
	printEncode(marshal.EncodeNat32(4294967295))
	// Output:
	// 7900000000
	// 792a000000
	// 79ffffffff
}

func ExampleEncodeNat64() {
	printEncode(marshal.EncodeNat64(0))
	printEncode(marshal.EncodeNat64(42))
	printEncode(marshal.EncodeNat64(1234567890))
	// Output:
	// 780000000000000000
	// 782a00000000000000
	// 78d202964900000000
}

func ExampleEncodeNat8() {
	printEncode(marshal.EncodeNat8(0))
	printEncode(marshal.EncodeNat8(42))
	printEncode(marshal.EncodeNat8(255))
	// Output:
	// 7b00
	// 7b2a
	// 7bff
}

func ExampleEncodeNull() {
	printEncode(marshal.EncodeNull())
	// Output:
	// 7f
}

func ExampleEncodePrincipal() {
	printEncode(marshal.EncodePrincipal(principalFromString("aaaaa-aa")))
	printEncode(marshal.EncodePrincipal(principalFromString("w7x7r-cok77-xa")))
	printEncode(marshal.EncodePrincipal(principalFromString("2chl6-4hpzw-vqaaa-aaaaa-c")))
	// Output:
	// 680100
	// 680103caffee
	// 680109efcdab000000000001
}

func ExampleEncodeText() {
	printEncode(marshal.EncodeText(""))
	printEncode(marshal.EncodeText("Motoko"))
	printEncode(marshal.EncodeText("Hi ☃\n"))
	// Output:
	// 7100
	// 71064d6f746f6b6f
	// 7107486920e298830a
}
