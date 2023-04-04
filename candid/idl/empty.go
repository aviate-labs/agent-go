package idl

import (
	"bytes"
	"fmt"
	"math/big"

	"github.com/aviate-labs/leb128"
)

// Empty represents the empty value.
type Empty struct{}

// EmptyType represents the empty type.
type EmptyType struct {
	primType
}

// Decode returns an error, as the empty type cannot be decoded.
func (EmptyType) Decode(*bytes.Reader) (any, error) {
	return nil, fmt.Errorf("cannot decode empty type")
}

// EncodeType returns the leb128 encoding of the empty type.
func (EmptyType) EncodeType(_ *TypeDefinitionTable) ([]byte, error) {
	return leb128.EncodeSigned(big.NewInt(emptyType))
}

// EncodeValue returns an empty byte slice.
func (EmptyType) EncodeValue(_ any) ([]byte, error) {
	return []byte{}, nil
}

// String returns the string representation of the type.
func (EmptyType) String() string {
	return "empty"
}
