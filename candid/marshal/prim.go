package marshal

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"math"
	"math/big"
	"unicode/utf8"

	"github.com/aviate-labs/agent-go/candid/idl"
	"github.com/aviate-labs/agent-go/principal"
	"github.com/aviate-labs/leb128"
)

func DecodeBool(r *bytes.Reader) (bool, error) {
	v, err := r.ReadByte()
	if err != nil {
		return false, err
	}
	switch v {
	case 0x00:
		return false, nil
	case 0x01:
		return true, nil
	default:
		return false, fmt.Errorf("invalid bool value: %v", v)
	}
}

func DecodeFloat32(r *bytes.Reader) (float32, error) {
	v := make([]byte, 4)
	if _, err := r.Read(v); err != nil {
		return 0, err
	}
	return math.Float32frombits(
		binary.LittleEndian.Uint32(v),
	), nil
}

func DecodeFloat64(r *bytes.Reader) (float64, error) {
	v := make([]byte, 8)
	if _, err := r.Read(v); err != nil {
		return 0, err
	}
	return math.Float64frombits(
		binary.LittleEndian.Uint64(v),
	), nil
}

func DecodeInt(r *bytes.Reader) (*big.Int, error) {
	return leb128.DecodeSigned(r)
}

func DecodeInt16(r *bytes.Reader) (int16, error) {
	bi, err := readInt(r, 2)
	if err != nil {
		return 0, err
	}
	return int16(bi.Int64()), nil
}

func DecodeInt32(r *bytes.Reader) (int32, error) {
	bi, err := readInt(r, 4)
	if err != nil {
		return 0, err
	}
	return int32(bi.Int64()), nil
}

func DecodeInt64(r *bytes.Reader) (int64, error) {
	bi, err := readInt(r, 8)
	if err != nil {
		return 0, err
	}
	return bi.Int64(), nil
}

func DecodeInt8(r *bytes.Reader) (int8, error) {
	bi, err := readInt(r, 1)
	if err != nil {
		return 0, err
	}
	return int8(bi.Int64()), nil
}

func DecodeNat(r *bytes.Reader) (*big.Int, error) {
	return leb128.DecodeUnsigned(r)
}

func DecodeNat16(r *bytes.Reader) (uint16, error) {
	v := make([]byte, 2)
	if _, err := r.Read(v); err != nil {
		return 0, err
	}
	return binary.LittleEndian.Uint16(v), nil
}

func DecodeNat32(r *bytes.Reader) (uint32, error) {
	v := make([]byte, 4)
	if _, err := r.Read(v); err != nil {
		return 0, err
	}
	return binary.LittleEndian.Uint32(v), nil
}

func DecodeNat64(r *bytes.Reader) (uint64, error) {
	v := make([]byte, 8)
	if _, err := r.Read(v); err != nil {
		return 0, err
	}
	return binary.LittleEndian.Uint64(v), nil
}

func DecodeNat8(r *bytes.Reader) (uint8, error) {
	return r.ReadByte()
}

func DecodePrincipal(r *bytes.Reader) (principal.Principal, error) {
	b, err := r.ReadByte()
	if err != nil {
		return principal.Principal{}, err
	}
	if b != 0x01 {
		return principal.Principal{}, fmt.Errorf("cannot decode principal")
	}
	l, err := leb128.DecodeUnsigned(r)
	if err != nil {
		return principal.Principal{}, err
	}
	if l.Uint64() == 0 {
		return principal.Principal{Raw: []byte{}}, nil
	}
	v := make([]byte, l.Uint64())
	if _, err := r.Read(v); err != nil {
		return principal.Principal{}, err
	}
	return principal.Principal{Raw: v}, nil
}

func DecodeText(r *bytes.Reader) (string, error) {
	n, err := leb128.DecodeUnsigned(r)
	if err != nil {
		return "", err
	}
	bs := make([]byte, n.Int64())
	if _, err := r.Read(bs); err != nil {
		return "", nil
	}
	if !utf8.Valid(bs) {
		return "", fmt.Errorf("invalid utf8 text")
	}
	return string(bs), nil
}

func EncodeBool(value bool) ([]byte, []byte, error) {
	var v []byte
	if v = []byte{0x00}; value {
		v = []byte{0x01}
	}
	return Bool.bytes(), v, nil
}

func EncodeEmpty() ([]byte, []byte, error) {
	return Empty.bytes(), []byte(nil), nil
}

func EncodeFloat32(value float32) ([]byte, []byte, error) {
	v := make([]byte, 4)
	binary.LittleEndian.PutUint32(v, math.Float32bits(value))
	return Float32.bytes(), v, nil
}

func EncodeFloat64(value float64) ([]byte, []byte, error) {
	v := make([]byte, 8)
	binary.LittleEndian.PutUint64(v, math.Float64bits(value))
	return Float64.bytes(), v, nil
}

func EncodeInt(value idl.Int) ([]byte, []byte, error) {
	v, err := leb128.EncodeSigned(value.BigInt())
	if err != nil {
		return nil, nil, err
	}
	return Int.bytes(), v, nil
}

func EncodeInt16(value int16) ([]byte, []byte, error) {
	return Int16.bytes(), writeInt(big.NewInt(int64(value)), 2), nil
}

func EncodeInt32(value int32) ([]byte, []byte, error) {
	return Int32.bytes(), writeInt(big.NewInt(int64(value)), 4), nil
}

func EncodeInt64(value int64) ([]byte, []byte, error) {
	return Int64.bytes(), writeInt(big.NewInt(value), 8), nil
}

func EncodeInt8(value int8) ([]byte, []byte, error) {
	return Int8.bytes(), writeInt(big.NewInt(int64(value)), 1), nil
}

func EncodeNat(value idl.Nat) ([]byte, []byte, error) {
	v, err := leb128.EncodeUnsigned(value.BigInt())
	if err != nil {
		return nil, nil, err
	}
	return Nat.bytes(), v, nil
}

func EncodeNat16(value uint16) ([]byte, []byte, error) {
	v := make([]byte, 2)
	binary.LittleEndian.PutUint16(v, value)
	return Nat16.bytes(), v, nil
}

func EncodeNat32(value uint32) ([]byte, []byte, error) {
	v := make([]byte, 4)
	binary.LittleEndian.PutUint32(v, value)
	return Nat32.bytes(), v, nil
}

func EncodeNat64(value uint64) ([]byte, []byte, error) {
	v := make([]byte, 8)
	binary.LittleEndian.PutUint64(v, value)
	return Nat64.bytes(), v, nil
}

func EncodeNat8(value uint8) ([]byte, []byte, error) {
	return Nat8.bytes(), []byte{value}, nil
}

func EncodeNull() ([]byte, []byte, error) {
	return Null.bytes(), []byte(nil), nil
}

func EncodePrincipal(value principal.Principal) ([]byte, []byte, error) {
	l, err := leb128.EncodeUnsigned(big.NewInt(int64(len(value.Raw))))
	if err != nil {
		return nil, nil, err
	}
	return Principal.bytes(), concat([]byte{0x01}, l, value.Raw), nil
}

func EncodeReserved() ([]byte, []byte, error) {
	return Reserved.bytes(), []byte(nil), nil
}

func EncodeText(value string) ([]byte, []byte, error) {
	l, err := leb128.EncodeUnsigned(big.NewInt(int64(len(value))))
	if err != nil {
		return nil, nil, err
	}
	return Text.bytes(), append(l, []byte(value)...), nil
}
