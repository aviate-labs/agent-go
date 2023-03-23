package idl

import (
	"bytes"
	"fmt"
	"math/big"

	"github.com/aviate-labs/leb128"
)

type BoolType struct {
	primType
}

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

func (BoolType) EncodeType(_ *TypeDefinitionTable) ([]byte, error) {
	return leb128.EncodeSigned(big.NewInt(boolType))
}

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

func (BoolType) String() string {
	return "bool"
}
