package idl

import (
	"bytes"
	"fmt"
	"math/big"

	"github.com/aviate-labs/leb128"
)

type Optional struct {
	V any
	T Type
}

func (o Optional) SubType() Type {
	return o.T
}

func (o Optional) Value() any {
	return o.V
}

type OptionalType struct {
	Type Type
}

func NewOptionalType(t Type) *OptionalType {
	return &OptionalType{
		Type: t,
	}
}

func (o OptionalType) AddTypeDefinition(tdt *TypeDefinitionTable) error {
	if err := o.Type.AddTypeDefinition(tdt); err != nil {
		return err
	}

	id, err := leb128.EncodeSigned(big.NewInt(optType))
	if err != nil {
		return err
	}
	v, err := o.Type.EncodeType(tdt)
	if err != nil {
		return err
	}
	tdt.Add(o, concat(id, v))
	return nil
}

func (o OptionalType) Decode(r *bytes.Reader) (any, error) {
	l, err := r.ReadByte()
	if err != nil {
		return nil, err
	}
	switch l {
	case 0x00:
		return nil, nil
	case 0x01:
		return o.Type.Decode(r)
	default:
		return nil, fmt.Errorf("invalid option value")
	}
}

func (o OptionalType) EncodeType(tdt *TypeDefinitionTable) ([]byte, error) {
	idx, ok := tdt.Indexes[o.String()]
	if !ok {
		return nil, fmt.Errorf("missing type index for: %s", o)
	}
	return leb128.EncodeSigned(big.NewInt(int64(idx)))
}

func (o OptionalType) EncodeValue(v any) ([]byte, error) {
	if v == nil {
		return []byte{0x00}, nil
	}
	v_, err := o.Type.EncodeValue(v)
	if err != nil {
		return nil, err
	}
	return concat([]byte{0x01}, v_), nil
}

func (o OptionalType) String() string {
	return fmt.Sprintf("opt %s", o.Type)
}

type OptionalValue interface {
	Value() any
	SubType() Type
}
