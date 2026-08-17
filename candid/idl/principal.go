package idl

import (
	"bytes"
	"fmt"
	"io"
	"math/big"

	"github.com/aviate-labs/agent-go/leb128"
	"github.com/aviate-labs/agent-go/principal"
)

type PrincipalType struct {
	primType
}

func (PrincipalType) Decode(r *bytes.Reader, budget *Budget) (any, error) {
	if err := budget.Spend(costValue); err != nil {
		return nil, err
	}
	b, err := r.ReadByte()
	if err != nil {
		return nil, err
	}
	if b != 0x01 {
		return nil, fmt.Errorf("cannot decode principal")
	}
	l, err := DecodeLen(r)
	if err != nil {
		return nil, err
	}
	if l == 0 {
		return principal.Principal{Raw: []byte{}}, nil
	}
	v := make([]byte, l)
	if _, err := io.ReadFull(r, v); err != nil {
		return nil, err
	}
	return principal.Principal{Raw: v}, nil
}

func (PrincipalType) EncodeType(_ *TypeDefinitionTable) ([]byte, error) {
	return leb128.EncodeSigned(PrincipalOpCode.BigInt())
}

func (PrincipalType) EncodeValue(v any) ([]byte, error) {
	v_, ok := v.(principal.Principal)
	if !ok {
		return nil, NewEncodeValueError(v, PrincipalOpCode)
	}
	l, err := leb128.EncodeUnsigned(big.NewInt(int64(len(v_.Raw))))
	if err != nil {
		return nil, err
	}
	return concat([]byte{0x01}, l, v_.Raw), nil
}

func (PrincipalType) Read(r *bytes.Reader) ([]byte, error) {
	b, err := r.ReadByte()
	if err != nil {
		return nil, err
	}
	if b != 0x01 {
		return nil, fmt.Errorf("cannot decode principal")
	}
	raw, err := readLEB128(r)
	if err != nil {
		return nil, err
	}
	l, err := leb128.DecodeUnsigned(bytes.NewReader(raw))
	if err != nil {
		return nil, err
	}
	if l.Uint64() == 0 {
		return append([]byte{b}, raw...), nil
	}
	n, err := checkLen(l, r)
	if err != nil {
		return nil, err
	}
	bs := make([]byte, 1+len(raw)+n)
	bs[0] = b
	copy(bs[1:], raw)
	if _, err := io.ReadFull(r, bs[1+len(raw):]); err != nil {
		return nil, err
	}
	return bs, nil
}

func (PrincipalType) String() string {
	return "principal"
}

func (PrincipalType) UnmarshalGo(raw any, _v any) error {
	v, ok := _v.(*principal.Principal)
	if !ok {
		return NewUnmarshalGoError(raw, _v)
	}
	if p, ok := raw.(principal.Principal); ok {
		*v = p
		return nil
	}
	if b, ok := raw.([]byte); ok {
		*v = principal.Principal{Raw: b}
		return nil
	}
	return NewUnmarshalGoError(raw, _v)
}
