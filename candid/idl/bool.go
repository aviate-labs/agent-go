package idl

import (
	"bytes"
	"fmt"
	"math/big"

	"github.com/aviate-labs/leb128"
)

// BoolType is a type of bool.
type BoolType struct {
	primType
}

// Decode decodes a bool value from the given reader.
func (b BoolType) Decode(r *bytes.Reader) (any, error) {
	v, err := r.ReadByte()
	if err != nil {
		return nil, err
	}
	switch v {
	case 0x00:
		return false, nil
	case 0x01:
		return true, nil
	default:
		return nil, fmt.Errorf("invalid bool values: %x", b)
	}
}

// EncodeType returns the leb128 encoding of the bool type.
func (BoolType) EncodeType(_ *TypeDefinitionTable) ([]byte, error) {
	return leb128.EncodeSigned(big.NewInt(boolType))
}

// EncodeValue encodes a bool value.
// Accepted types are: `bool`.
func (BoolType) EncodeValue(v any) ([]byte, error) {
	v_, ok := v.(bool)
	if !ok {
		return nil, NewEncodeValueError(v, boolType)
	}
	if v_ {
		return []byte{0x01}, nil
	}
	return []byte{0x00}, nil
}

// String returns the string representation of the type.
func (BoolType) String() string {
	return "bool"
}

func (BoolType) UnmarshalGo(raw any, _v any) error {
	v, ok := _v.(*bool)
	if !ok {
		return NewUnmarshalGoError(raw, _v)
	}
	b, ok := raw.(bool)
	if !ok {
		return NewUnmarshalGoError(raw, _v)
	}
	*v = b
	return nil
}
