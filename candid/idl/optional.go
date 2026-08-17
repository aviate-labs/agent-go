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

// Some marks a decoded option as present. It only wraps values that would
// otherwise be indistinguishable from absence: `opt null` and `null` both
// decode to a bare nil without it.
type Some struct {
	Value any
}

// Decode decodes the value from the given reader into either `nil` or a value (of the subtype of the optional type).
// A present option whose value is itself nil is returned as Some, so that
// `opt null` stays distinguishable from `null`.
func (o OptionalType) Decode(r *bytes.Reader, budget *Budget) (any, error) {
	if err := budget.Spend(costComposite); err != nil {
		return nil, err
	}
	b, err := r.ReadByte()
	if err != nil {
		return nil, err
	}
	switch b {
	case 0x00:
		return nil, nil
	case 0x01:
		v, err := o.Type.Decode(r, budget)
		if err != nil {
			return nil, err
		}
		if v == nil {
			return Some{}, nil
		}
		return v, nil
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
		// Unwrap one level only: each pointer level corresponds to one opt, so
		// recursing on o here would let the outer opt swallow the inner one and
		// encode `opt null` the same as `null`.
		v_, err := o.Type.EncodeValue(v.Elem().Interface())
		if err != nil {
			return nil, err
		}
		return append([]byte{0x01}, v_...), nil
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
	return typeString(&o, nil)
}

func (o *OptionalType) elided() string { return "opt" }

func (o *OptionalType) stringSeen(seen []Type) string {
	return fmt.Sprintf("opt %s", typeString(o.Type, seen))
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
		// A present option carrying null: allocate the outer pointer but leave
		// the inner one nil, without descending.
		if s, ok := raw.(Some); ok {
			ptr := reflect.New(v.Type().Elem())
			if s.Value != nil {
				if err := UnmarshalGo(o.Type, s.Value, ptr.Interface()); err != nil {
					v.Set(reflect.Zero(v.Type()))
					return nil
				}
			}
			v.Set(ptr)
			return nil
		}
		// Any type is a subtype of an option: a value the receiver cannot
		// interpret (an added variant tag, a changed constituent type) is
		// seen as null rather than being an error.
		ptr := reflect.New(v.Type().Elem())
		if err := UnmarshalGo(o.Type, raw, ptr.Interface()); err != nil {
			v.Set(reflect.Zero(v.Type()))
			return nil
		}
		// A nested opt degrades in its own frame and reports success. Absent a
		// Some marker the value was not a received null, so a nil here means the
		// child dropped it: collapse rather than wrap it in a non-nil pointer.
		if _, nested := o.Type.(*OptionalType); nested && ptr.Elem().IsNil() {
			v.Set(reflect.Zero(v.Type()))
			return nil
		}
		v.Set(ptr)
		return nil
	}
	// Nothing to assign to v.
	return NewUnmarshalGoError(raw, _v)
}
