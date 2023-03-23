package idl

import (
	"bytes"
	"fmt"
	"math/big"

	"github.com/aviate-labs/leb128"
)

type VectorType struct {
	Type Type
}

func NewVectorType(t Type) *VectorType {
	return &VectorType{
		Type: t,
	}
}

func (vec VectorType) AddTypeDefinition(tdt *TypeDefinitionTable) error {
	if err := vec.Type.AddTypeDefinition(tdt); err != nil {
		return err
	}

	id, err := leb128.EncodeSigned(big.NewInt(vecType))
	if err != nil {
		return err
	}
	v_, err := vec.Type.EncodeType(tdt)
	if err != nil {
		return err
	}
	tdt.Add(vec, concat(id, v_))
	return nil
}

func (vec VectorType) Decode(r *bytes.Reader) (any, error) {
	l, err := leb128.DecodeUnsigned(r)
	if err != nil {
		return nil, err
	}
	var vs []any
	for i := 0; i < int(l.Int64()); i++ {
		v_, err := vec.Type.Decode(r)
		if err != nil {
			return nil, err
		}
		vs = append(vs, v_)
	}
	return vs, nil
}

func (vec VectorType) EncodeType(tdt *TypeDefinitionTable) ([]byte, error) {
	idx, ok := tdt.Indexes[vec.String()]
	if !ok {
		return nil, fmt.Errorf("missing type index for: %s", vec)
	}
	return leb128.EncodeSigned(big.NewInt(int64(idx)))
}

func (vec VectorType) EncodeValue(v any) ([]byte, error) {
	vs_, ok := v.([]any)
	if !ok {
		return nil, NewEncodeValueError(v, vecType)
	}
	l, err := leb128.EncodeSigned(big.NewInt(int64(len(vs_))))
	if err != nil {
		return nil, err
	}
	var vs []byte
	for _, value := range vs_ {
		v_, err := vec.Type.EncodeValue(value)
		if err != nil {
			return nil, err
		}
		vs = append(vs, v_...)
	}
	return concat(l, vs), nil
}

func (vec VectorType) String() string {
	return fmt.Sprintf("vec %s", vec.Type)
}
