package idl

import (
	"bytes"
	"fmt"
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
		return nil, fmt.Errorf("invalid argument: %v", v)
	}
	return []byte{}, nil
}

func (NullType) String() string {
	return "null"
}
