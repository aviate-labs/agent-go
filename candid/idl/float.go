package idl

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"github.com/aviate-labs/leb128"
	"math"
	"math/big"
)

func anyToFloat64(v any) (float64, bool) {
	switch v := v.(type) {
	case float64:
		return v, true
	case float32:
		return float64(v), true
	default:
		return 0, false
	}
}

// encodeFloat32 convert the given value to a float32.
// Accepts: `float32`.
func encodeFloat32(v any) (float32, error) {
	if v, ok := v.(float32); ok {
		return v, nil
	}
	return 0, fmt.Errorf("invalid value: %v", v)
}

// encodeFloat64 convert the given value to a float64.
// Accepts: `float32`, `float64`.
func encodeFloat64(v any) (float64, error) {
	if v, ok := v.(float64); ok {
		return v, nil
	}
	v_, err := encodeFloat32(v)
	return float64(v_), err
}

// FloatType is either a type of float32 or float64.
// Should only be initialized through `Float32Type` and `Float64Type`.
type FloatType struct {
	size uint8
	primType
}

// Float32Type returns a type of float32.
func Float32Type() *FloatType {
	return &FloatType{
		size: 4,
	}
}

// Float64Type returns a type of float64.
func Float64Type() *FloatType {
	return &FloatType{
		size: 8,
	}
}

// Base returns the base type of the float type.
// Either `4` (32) or `8` (64).
func (f FloatType) Base() uint {
	return uint(f.size)
}

// Decode decodes a float value from the given reader.
func (f FloatType) Decode(r *bytes.Reader) (any, error) {
	switch f.size {
	case 4:
		v := make([]byte, f.size)
		n, err := r.Read(v)
		if err != nil {
			return nil, err
		}
		if uint8(n) != f.size {
			return nil, fmt.Errorf("float32: too short")
		}
		return math.Float32frombits(
			binary.LittleEndian.Uint32(v),
		), nil
	case 8:
		v := make([]byte, f.size)
		n, err := r.Read(v)
		if err != nil {
			return nil, err
		}
		if uint8(n) != f.size {
			return nil, fmt.Errorf("float64: too short")
		}
		return math.Float64frombits(
			binary.LittleEndian.Uint64(v),
		), nil
	default:
		return nil, fmt.Errorf("invalid float type with size %d", f.size)
	}
}

// EncodeType returns the leb128 encoding of the FloatType.
func (f FloatType) EncodeType(_ *TypeDefinitionTable) ([]byte, error) {
	floatXType := new(big.Int).Set(big.NewInt(floatXType))
	if f.size == 8 {
		floatXType.Add(floatXType, big.NewInt(-1))
	}
	return leb128.EncodeSigned(floatXType)
}

// EncodeValue encodes a float value.
// Accepted types are: `float32` and `float64`.
func (f FloatType) EncodeValue(v any) ([]byte, error) {
	switch f.size {
	case 8:
		v, err := encodeFloat64(v)
		if err != nil {
			return nil, err
		}
		bs := make([]byte, f.size)
		binary.LittleEndian.PutUint64(bs, math.Float64bits(v))
		return bs, nil
	case 4:
		v, err := encodeFloat32(v)
		if err != nil {
			return nil, err
		}
		bs := make([]byte, f.size)
		binary.LittleEndian.PutUint32(bs, math.Float32bits(v))
		return bs, nil
	default:
		return nil, NewEncodeValueError(v, floatXType)
	}
}

// String returns the string representation of the type.
func (f FloatType) String() string {
	return fmt.Sprintf("float%d", f.size*8)
}

func (f FloatType) UnmarshalGo(raw any, _v any) error {
	switch f.size {
	case 8:
		f, ok := anyToFloat64(raw)
		if !ok {
			return NewUnmarshalGoError(raw, _v)
		}
		v, ok := _v.(*float64)
		if !ok {
			return NewUnmarshalGoError(raw, _v)
		}
		*v = f
		return nil
	case 4:
		f32, ok := raw.(float32)
		if !ok {
			return NewUnmarshalGoError(raw, _v)
		}
		switch v := _v.(type) {
		case *float32:
			*v = f32
			return nil
		case *float64:
			*v = float64(f32)
			return nil
		default:
			return NewUnmarshalGoError(raw, _v)
		}
	default:
		return NewUnmarshalGoError(raw, _v)
	}
}
