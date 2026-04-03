package idl

import (
	"bytes"
	"fmt"
	"io"
	"unicode/utf8"

	"github.com/niccolofant/agent-go/leb128"
)

// TextType is the type of a text value.
type TextType struct {
	primType
}

// Decode decodes the value from the given reader into a string.
func (TextType) Decode(r *bytes.Reader) (any, error) {
	n, err := decodeLen(r)
	if err != nil {
		return nil, err
	}
	if n == 0 {
		return "", nil
	}
	bs := make([]byte, n)
	i, err := r.Read(bs)
	if err != nil {
		return nil, err
	}
	if i != n {
		return nil, io.EOF
	}
	if !utf8.Valid(bs) {
		return nil, fmt.Errorf("invalid utf8 text: %s", string(bs))
	}
	return string(bs), nil
}

// EncodeType encodes the type into a byte slice.
func (TextType) EncodeType(_ *TypeDefinitionTable) ([]byte, error) {
	return leb128.EncodeSigned(TextOpCode.BigInt())
}

// EncodeValue encodes the value into a byte slice.
func (TextType) EncodeValue(v any) ([]byte, error) {
	v_, ok := v.(string)
	if !ok {
		return nil, NewEncodeValueError(v, TextOpCode)
	}
	var buf [10]byte
	bs := leb128.AppendUnsignedUint64(buf[:0], uint64(len(v_)))
	return append(bs, v_...), nil
}

func (TextType) Read(r *bytes.Reader) ([]byte, error) {
	raw, err := readLEB128(r)
	if err != nil {
		return nil, err
	}
	n, err := leb128.DecodeUnsigned(bytes.NewReader(raw))
	if err != nil {
		return nil, err
	}
	if n.Int64() == 0 {
		return raw, nil
	}
	bs := make([]byte, len(raw)+int(n.Int64()))
	i, err := r.Read(bs[len(raw)+1:])
	if err != nil {
		return nil, err
	}
	if i != int(n.Int64()) {
		return nil, io.EOF
	}
	return bs, nil
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
