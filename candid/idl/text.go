package idl

import (
	"bytes"
	"fmt"
	"io"
	"math/big"
	"unicode/utf8"

	"github.com/aviate-labs/leb128"
)

// TextType is the type of a text value.
type TextType struct {
	primType
}

// Decode decodes the value from the given reader into a string.
func (TextType) Decode(r *bytes.Reader) (any, error) {
	n, err := leb128.DecodeUnsigned(r)
	if err != nil {
		return nil, err
	}
	bs := make([]byte, n.Int64())
	i, err := r.Read(bs)
	if err != nil {
		return "", nil
	}
	if i != int(n.Int64()) {
		return nil, io.EOF
	}
	if !utf8.Valid(bs) {
		return nil, fmt.Errorf("invalid utf8 text")
	}

	return string(bs), nil
}

// EncodeType encodes the type into a byte slice.
func (TextType) EncodeType(_ *TypeDefinitionTable) ([]byte, error) {
	return leb128.EncodeSigned(big.NewInt(textType))
}

// EncodeValue encodes the value into a byte slice.
func (TextType) EncodeValue(v any) ([]byte, error) {
	v_, ok := v.(string)
	if !ok {
		return nil, NewEncodeValueError(v, textType)
	}
	bs, err := leb128.EncodeUnsigned(big.NewInt(int64(len(v_))))
	if err != nil {
		return nil, err
	}
	return append(bs, []byte(v_)...), nil
}

// String returns the string representation of the type.
func (TextType) String() string {
	return "text"
}

func (t TextType) UnmarshalGo(raw any, _v any) error {
	v, ok := _v.(*string)
	if !ok {
		return NewUnmarshalGoError(raw, _v)
	}
	b, ok := raw.(string)
	if !ok {
		return NewUnmarshalGoError(raw, _v)
	}
	*v = b
	return nil
}
