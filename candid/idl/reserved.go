package idl

import (
	"bytes"

	"github.com/niccolofant/agent-go/leb128"
)

type Reserved struct{}

type ReservedType struct {
	primType
}

func (ReservedType) Decode(*bytes.Reader) (any, error) {
	return nil, nil
}

func (ReservedType) EncodeType(_ *TypeDefinitionTable) ([]byte, error) {
	return leb128.EncodeSigned(ReservedOpCode.BigInt())
}

func (ReservedType) EncodeValue(_ any) ([]byte, error) {
	return []byte{}, nil
}

func (ReservedType) Read(*bytes.Reader) ([]byte, error) {
	return nil, nil
}

func (ReservedType) String() string {
	return "reserved"
}

func (ReservedType) UnmarshalGo(raw any, _v any) error {
	return NewUnmarshalGoError(raw, _v)
}
