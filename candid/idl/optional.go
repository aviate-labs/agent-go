package idl

import (
	"bytes"
	"fmt"
	"github.com/aviate-labs/leb128"
	"math/big"
	"reflect"
)

// Optional is a type that can be either nil or a value of the given type.
type Optional struct {
	V any
	T Type
}

// Subtype returns the type of the value.
func (o Optional) Subtype() Type {
	return o.T
}

// Value returns the value.
func (o Optional) Value() any {
	return o.V
}

// OptionalType is the type of an optional value.
type OptionalType struct {
	Type Type
}

// NewOptionalType creates a new optional type.
func NewOptionalType(t Type) *OptionalType {
	return &OptionalType{
		Type: t,
	}
}

// AddTypeDefinition adds the type definition to the table.
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

// Decode decodes the value from the given reader into either `nil` or a value (of the subtype of the optional type).
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

// EncodeType encodes the type into a byte array.
func (o OptionalType) EncodeType(tdt *TypeDefinitionTable) ([]byte, error) {
	idx, ok := tdt.Indexes[o.String()]
	if !ok {
		return nil, fmt.Errorf("missing type index for: %v", o)
	}
	return leb128.EncodeSigned(big.NewInt(int64(idx)))
}

// EncodeValue encodes the value into a byte array.
// Accepts `nil` or a value (of the subtype of the optional type).
func (o OptionalType) EncodeValue(v any) ([]byte, error) {
	if v == nil {
		return []byte{0x00}, nil
	}
	if v, ok := v.(OptionalValue); ok {
		return o.EncodeValue(v.Value())
	}
	if v := reflect.ValueOf(v); v.Kind() == reflect.Ptr {
		return o.EncodeValue(v.Elem().Interface())
	}
	v_, err := o.Type.EncodeValue(v)
	if err != nil {
		return nil, err
	}
	return concat([]byte{0x01}, v_), nil
}

// String returns the string representation of the type.
func (o OptionalType) String() string {
	return fmt.Sprintf("opt %s", o.Type)
}

// OptionalValue is a value of an optional type.
type OptionalValue interface {
	// Value returns the value.
	Value() any
	// Subtype returns the type of the value.
	Subtype() Type
}
