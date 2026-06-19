package idl

import (
	"bytes"
	"fmt"
	"math/big"
	"reflect"

	"github.com/aviate-labs/agent-go/leb128"
)

//go:fix inline
func Ptr[a any](v a) *a {
	return new(v)
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

	id, err := leb128.EncodeSigned(OptOpCode.BigInt())
	if err != nil {
		return err
	}
	v, err := o.Type.EncodeType(tdt)
	if err != nil {
		return err
	}
	tdt.Add(o, append(id, v...))
	return nil
}

// Decode decodes the value from the given reader into either `nil` or a value (of the subtype of the optional type).
func (o OptionalType) Decode(r *bytes.Reader) (any, error) {
	b, err := r.ReadByte()
	if err != nil {
		return nil, err
	}
	switch b {
	case 0x00:
		return nil, nil
	case 0x01:
		return o.Type.Decode(r)
	default:
		return nil, fmt.Errorf("invalid option value: %x", b)
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
	if v := reflect.ValueOf(v); v.Kind() == reflect.Pointer {
		if v.IsNil() {
			return []byte{0x00}, nil
		}
		return o.EncodeValue(v.Elem().Interface())
	}
	v_, err := o.Type.EncodeValue(v)
	if err != nil {
		return nil, err
	}
	return append([]byte{0x01}, v_...), nil
}

func (o OptionalType) Read(r *bytes.Reader) ([]byte, error) {
	b, err := r.ReadByte()
	if err != nil {
		return nil, err
	}
	switch b {
	case 0x00:
		return []byte{b}, nil
	case 0x01:
		bs, err := o.Type.Read(r)
		if err != nil {
			return nil, err
		}
		return append([]byte{b}, bs...), nil
	default:
		return nil, fmt.Errorf("invalid option value: %x", b)
	}
}

// String returns the string representation of the type.
func (o OptionalType) String() string {
	return fmt.Sprintf("opt %s", o.Type)
}

func (o OptionalType) UnmarshalGo(raw any, _v any) error {
	if raw == nil {
		// Optional value is `nil`.
		return nil
	}
	if v := reflect.ValueOf(_v); v.Kind() == reflect.Pointer {
		v := v.Elem() // Dereference the pointer.
		if k := v.Kind(); k != reflect.Pointer {
			return NewUnmarshalGoError(raw, _v)
		}
		if !v.IsNil() {
			// No need to allocate a new pointer.
			return UnmarshalGo(o.Type, raw, v.Interface())
		}
		ptr := reflect.New(v.Type().Elem()) // Create a new pointer.
		if err := UnmarshalGo(o.Type, raw, ptr.Interface()); err != nil {
			return err
		}
		v.Set(ptr)
		return nil
	}
	// Nothing to assign to v.
	return NewUnmarshalGoError(raw, _v)
}
