package idl

import (
	"bytes"
	"math/big"

	"github.com/aviate-labs/leb128"
)

type Null struct{}

type NullType struct {
	primType
}

func (NullType) Decode(_ *bytes.Reader) (any, error) {
	return nil, nil
}

func (NullType) EncodeType(_ *TypeDefinitionTable) ([]byte, error) {
	return leb128.EncodeSigned(big.NewInt(nullType))
}

func (NullType) EncodeValue(v any) ([]byte, error) {
	if v != nil {
		return nil, NewEncodeValueError(v, nullType)
	}
	return []byte{}, nil
}

func (NullType) String() string {
	return "null"
}
