package idl

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"math"
	"math/big"
	"reflect"

	"github.com/aviate-labs/leb128"
)

type FloatType struct {
	size uint8
	primType
}

func Float32Type() *FloatType {
	return &FloatType{
		size: 4,
	}
}

func Float64Type() *FloatType {
	return &FloatType{
		size: 8,
	}
}

func (f FloatType) Base() uint {
	return uint(f.size)
}

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

func (f FloatType) EncodeType(_ *TypeDefinitionTable) ([]byte, error) {
	floatXType := new(big.Int).Set(big.NewInt(floatXType))
	if f.size == 8 {
		floatXType.Add(floatXType, big.NewInt(-1))
	}
	return leb128.EncodeSigned(floatXType)
}

func (f FloatType) EncodeValue(v any) ([]byte, error) {
	return encode(reflect.ValueOf(v), func(k reflect.Kind, v reflect.Value) ([]byte, error) {
		switch k {
		case reflect.Float32:
			bs := make([]byte, f.size)
			binary.LittleEndian.PutUint32(bs, math.Float32bits(float32(v.Float())))
			return bs, nil
		case reflect.Float64:
			if f.size == 4 {
				return nil, fmt.Errorf("can not encode float64 into float32")
			}
			bs := make([]byte, f.size)
			binary.LittleEndian.PutUint64(bs, math.Float64bits(float64(v.Float())))
			return bs, nil
		default:
			return nil, fmt.Errorf("invalid float value: %s", v.Kind())
		}
	})
}

func (f FloatType) String() string {
	return fmt.Sprintf("float%d", f.size*8)
}
