package idl

import (
	"bytes"
	"github.com/aviate-labs/leb128"
	"math/big"
	"reflect"
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
	if _, ok := v.(Null); !ok && v != nil {
		return nil, NewEncodeValueError(v, nullType)
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
