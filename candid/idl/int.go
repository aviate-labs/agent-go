package idl

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"math/big"

	"github.com/aviate-labs/leb128"
)

func anyToInt16(v any) (int16, bool) {
	switch v := v.(type) {
	case int16:
		return v, true
	case int8:
		return int16(v), true
	default:
		return 0, false
	}
}

func anyToInt32(v any) (int32, bool) {
	switch v := v.(type) {
	case int32:
		return v, true
	case int16:
		return int32(v), true
	case int8:
		return int32(v), true
	default:
		return 0, false
	}
}

func anyToInt64(v any) (int64, bool) {
	switch v := v.(type) {
	case int64:
		return v, true
	case int32:
		return int64(v), true
	case int16:
		return int64(v), true
	case int8:
		return int64(v), true
	default:
		return 0, false
	}
}

// encodeInt16 convert the given value to an int16.
// Accepts: `int8`, `int16`.
func encodeInt16(v any) (int16, error) {
	if v, ok := v.(int16); ok {
		return v, nil
	}
	v_, err := encodeInt8(v)
	return int16(v_), err
}

// encodeInt32 convert the given value to an int32.
// Accepts: `int8`, `int16`, `int32`.
func encodeInt32(v any) (int32, error) {
	if v, ok := v.(int32); ok {
		return v, nil
	}
	v_, err := encodeInt16(v)
	return int32(v_), err
}

// encodeInt64 convert the given value to an int64.
// Accepts: `int8`, `int16`, `int32`, `int64`.
func encodeInt64(v any) (int64, error) {
	if v, ok := v.(int); ok {
		return int64(v), nil
	}
	if v, ok := v.(int64); ok {
		return v, nil
	}
	v_, err := encodeInt16(v)
	return int64(v_), err
}

// encodeInt8 convert the given value to an int8.
// Accepts: `int8`.
func encodeInt8(v any) (int8, error) {
	if v, ok := v.(int8); ok {
		return v, nil
	}
	return 0, fmt.Errorf("invalid value: %v", v)
}

// Int represents an unbounded integer.
type Int struct {
	i *big.Int
}

// NewBigInt creates a new Int from a big.Int.
func NewBigInt(bi *big.Int) Int {
	return Int{bi}
}

// NewInt creates a new Int from any integer.
func NewInt[number Integer](i number) Int {
	return Int{i: big.NewInt(int64(i))}
}

// NewIntFromString creates a new Int from a string.
func NewIntFromString(n string) Int {
	bi, ok := new(big.Int).SetString(n, 10)
	if !ok {
		panic("number: invalid string: " + n)
	}
	return Int{bi}
}

func anyToInt(v any) (Int, bool) {
	switch v := v.(type) {
	case Int:
		return v, true
	case int:
		return NewInt(v), true
	case int64:
		return NewInt(v), true
	case int32:
		return NewInt(v), true
	case int16:
		return NewInt(v), true
	case int8:
		return NewInt(v), true
	default:
		return Int{}, false
	}
}

// BigInt returns the underlying big.Int.
func (i Int) BigInt() *big.Int {
	return i.i
}

// String returns the string representation of the Int.
func (i Int) String() string {
	return i.i.String()
}

// IntType is either a type of int8, int16, int32, int64, or int.
type IntType struct {
	size uint8
	primType
}

// Int16Type returns a type of int16.
func Int16Type() *IntType {
	return &IntType{
		size: 2,
	}
}

// Int32Type returns a type of int32.
func Int32Type() *IntType {
	return &IntType{
		size: 4,
	}
}

// Int64Type returns a type of int64.
func Int64Type() *IntType {
	return &IntType{
		size: 8,
	}
}

// Int8Type returns a type of int8.
func Int8Type() *IntType {
	return &IntType{
		size: 1,
	}
}

// Base returns the base type of the IntType.
func (n IntType) Base() uint {
	return uint(n.size)
}

// Decode decodes an integer from the given reader.
func (n IntType) Decode(r *bytes.Reader) (any, error) {
	switch n.size {
	case 0:
		bi, err := leb128.DecodeSigned(r)
		if err != nil {
			return nil, err
		}
		return NewBigInt(bi), nil
	case 8:
		v := make([]byte, 8)
		n, err := r.Read(v)
		if err != nil {
			return nil, err
		}
		if n != 8 {
			return nil, fmt.Errorf("int64: too short")
		}
		bi, err := readInt(new(big.Int).SetUint64(binary.LittleEndian.Uint64(v)), 8)
		if err != nil {
			return nil, err
		}
		return bi.Int64(), nil
	case 4:
		v := make([]byte, 4)
		n, err := r.Read(v)
		if err != nil {
			return nil, err
		}
		if n != 4 {
			return nil, fmt.Errorf("int32: too short")
		}
		bi, err := readInt(new(big.Int).SetUint64(uint64(binary.LittleEndian.Uint32(v))), 8)
		if err != nil {
			return nil, err
		}
		return int32(bi.Int64()), nil
	case 2:
		v := make([]byte, 2)
		n, err := r.Read(v)
		if err != nil {
			return nil, err
		}
		if n != 2 {
			return nil, fmt.Errorf("int16: too short")
		}
		bi, err := readInt(new(big.Int).SetUint64(uint64(binary.LittleEndian.Uint16(v))), 8)
		if err != nil {
			return nil, err
		}
		return int16(bi.Int64()), nil
	case 1:
		v, err := r.ReadByte()
		if err != nil {
			return nil, err
		}
		bi, err := readInt(new(big.Int).SetUint64(uint64(v)), 8)
		if err != nil {
			return nil, err
		}
		return int8(bi.Int64()), nil
	default:
		return nil, fmt.Errorf("invalid int type with size %d", n.size)
	}
}

// EncodeType returns the leb128 encoding of the IntType.
func (n IntType) EncodeType(_ *TypeDefinitionTable) ([]byte, error) {
	if n.size == 0 {
		return leb128.EncodeSigned(big.NewInt(intType))
	}
	intXType := new(big.Int).Set(big.NewInt(intXType))
	intXType = intXType.Add(
		intXType,
		big.NewInt(3-int64(log2(n.size*8))),
	)
	return leb128.EncodeSigned(intXType)
}

// EncodeValue encodes an int value.
// Accepted types are: `int`, `int8`, `int16`, `int32`, `int64`, `Int`.
func (n IntType) EncodeValue(v any) ([]byte, error) {
	switch n.size {
	case 0:
		v, ok := v.(Int)
		if !ok {
			return nil, fmt.Errorf("invalid value: %v", v)
		}
		return leb128.EncodeSigned(v.BigInt())
	case 8:
		v, err := encodeInt64(v)
		if err != nil {
			return nil, err
		}
		return writeInt(big.NewInt(v), 8), nil
	case 4:
		v, err := encodeInt32(v)
		if err != nil {
			return nil, err
		}
		return writeInt(big.NewInt(int64(v)), 4), nil
	case 2:
		v, err := encodeInt16(v)
		if err != nil {
			return nil, err
		}
		return writeInt(big.NewInt(int64(v)), 2), nil
	case 1:
		v, err := encodeInt8(v)
		if err != nil {
			return nil, err
		}
		return writeInt(big.NewInt(int64(v)), 1), nil
	default:
		return nil, NewEncodeValueError(v, intType)
	}
}

// String returns the string representation of the type.
func (n IntType) String() string {
	if n.size == 0 {
		return "int"
	}
	return fmt.Sprintf("int%d", n.size*8)
}

func (n IntType) UnmarshalGo(raw any, _v any) error {
	switch n.size {
	case 0:
		n, ok := anyToInt(raw)
		if !ok {
			return NewUnmarshalGoError(raw, _v)
		}
		v, ok := _v.(*Int)
		if !ok {
			return NewUnmarshalGoError(raw, _v)
		}
		*v = n
		return nil
	case 8:
		i64, ok := anyToInt64(raw)
		if !ok {
			return NewUnmarshalGoError(raw, _v)
		}
		switch v := _v.(type) {
		case *int64:
			*v = i64
			return nil
		case *Int:
			*v = NewInt(i64)
			return nil
		default:
			return NewUnmarshalGoError(raw, _v)
		}
	case 4:
		i32, ok := anyToInt32(raw)
		if !ok {
			return NewUnmarshalGoError(raw, _v)
		}
		switch v := _v.(type) {
		case *int32:
			*v = i32
			return nil
		case *Int:
			*v = NewInt(i32)
			return nil
		case *int64:
			*v = int64(i32)
			return nil
		default:
			return NewUnmarshalGoError(raw, _v)
		}
	case 2:
		i16, ok := anyToInt16(raw)
		if !ok {
			return NewUnmarshalGoError(raw, _v)
		}
		switch v := _v.(type) {
		case *int16:
			*v = i16
			return nil
		case *Int:
			*v = NewInt(i16)
			return nil
		case *int64:
			*v = int64(i16)
			return nil
		case *int32:
			*v = int32(i16)
			return nil
		default:
			return NewUnmarshalGoError(raw, _v)
		}
	case 1:
		i8, ok := raw.(int8)
		if !ok {
			return NewUnmarshalGoError(raw, _v)
		}
		switch v := _v.(type) {
		case *int8:
			*v = i8
			return nil
		case *Int:
			*v = NewInt(i8)
			return nil
		case *int64:
			*v = int64(i8)
			return nil
		case *int32:
			*v = int32(i8)
			return nil
		case *int16:
			*v = int16(i8)
			return nil
		default:
			return NewUnmarshalGoError(raw, _v)
		}
	default:
		return NewUnmarshalGoError(raw, _v)
	}
}

// Integer contains all integer types.
type Integer interface {
	int | int64 | int32 | int16 | int8
}
