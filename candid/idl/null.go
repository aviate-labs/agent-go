package idl

import (
	"bytes"
	"reflect"

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
	return leb128.EncodeSigned(NullOpCode.BigInt())
}

func (NullType) EncodeValue(v any) ([]byte, error) {
	if _, ok := v.(Null); !ok && v != nil {
		return nil, NewEncodeValueError(v, NullOpCode)
	}
	return []byte{}, nil
}

func (NullType) String() string {
	return "null"
}

func (NullType) UnmarshalGo(raw any, _v any) error {
	v := reflect.ValueOf(_v)
	if v.Kind() != reflect.Ptr {
		return NewUnmarshalGoError(raw, _v)
	}
	if _, ok := raw.(Null); ok || raw == nil {
		return nil
	}
	return NewUnmarshalGoError(raw, _v)
}
