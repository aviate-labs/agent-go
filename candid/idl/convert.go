package idl

import (
	"fmt"
	"reflect"

	"github.com/aviate-labs/agent-go/principal"
)

func EmptyOf(t Type) (any, error) {
	var v any = t
	if r := reflect.ValueOf(t); r.Kind() == reflect.Pointer {
		v = r.Elem().Interface()
	}

	switch t := v.(type) {
	case NullType:
		return Null{}, nil
	case BoolType:
		return false, nil
	case NatType:
		switch t.size {
		case 1:
			return uint8(0), nil
		case 2:
			return uint16(0), nil
		case 4:
			return uint32(0), nil
		case 8:
			return uint64(0), nil
		default:
			return NewNat(uint(0)), nil
		}
	case IntType:
		switch t.size {
		case 1:
			return int8(0), nil
		case 2:
			return int16(0), nil
		case 4:
			return int32(0), nil
		case 8:
			return int64(0), nil
		default:
			return NewInt(0), nil
		}
	case FloatType:
		switch t.size {
		case 4:
			return float32(0), nil
		case 8:
			return float64(0), nil
		}
	case TextType:
		return "", nil
	case ReservedType:
		return Reserved{}, nil
	case EmptyType:
		return Empty{}, nil
	case PrincipalType:
		return principal.Principal{}, nil
	case OptionalType:
		if v, err := EmptyOf(t.Type); err == nil {
			return &v, nil
		}
	case VectorType:
		if v, err := EmptyOf(t.Type); err == nil {
			return []any{v}, nil
		}
	case RecordType:
		fields := make(map[string]any)
		for _, field := range t.Fields {
			v, err := EmptyOf(field.Type)
			if err != nil {
				return nil, UnknownTypeError{Type: t}
			}
			fields[field.Name] = v
		}
		return fields, nil
	case VariantType:
		if len(t.Fields) == 0 {
			return Variant{
				Type: t,
			}, nil
		}
		field := t.Fields[0]
		if v, err := EmptyOf(field.Type); err == nil {
			return Variant{
				Name:  field.Name,
				Value: v,
				Type:  t,
			}, nil
		}
	}
	return nil, UnknownTypeError{Type: t}
}

func TypeOf(v any) (Type, error) {
	switch v := v.(type) {
	case Null:
		return new(NullType), nil
	case bool:
		return new(BoolType), nil
	case Nat:
		return new(NatType), nil
	case Int:
		return new(IntType), nil
	case uint8:
		return Nat8Type(), nil
	case uint16:
		return Nat16Type(), nil
	case uint32:
		return Nat32Type(), nil
	case uint, uint64:
		return Nat64Type(), nil
	case int8:
		return Int8Type(), nil
	case int16:
		return Int16Type(), nil
	case int32:
		return Int32Type(), nil
	case int, int64:
		return Int64Type(), nil
	case float32:
		return Float32Type(), nil
	case float64:
		return Float64Type(), nil
	case string:
		return new(TextType), nil
	case Reserved:
		return new(ReservedType), nil
	case Empty:
		return new(EmptyType), nil
	case []any:
		typ, err := TypeOf(v[0])
		if err != nil {
			return nil, err
		}
		return NewVectorType(typ), nil
	case map[string]any:
		fields := make(map[string]Type)
		for k, v := range v {
			typ, err := TypeOf(v)
			if err != nil {
				return nil, err
			}
			fields[k] = typ
		}
		return NewRecordType(fields), nil
	case Variant:
		fields := make(map[string]Type)
		typ, err := TypeOf(v.Value)
		if err != nil {
			return nil, err
		}
		fields[v.Name] = typ
		return NewVariantType(fields), nil
	case principal.Principal:
		return new(PrincipalType), nil
	default:
		if v == nil {
			return new(NullType), nil
		}

		// Derive the type from the reflect.Type rather than from fabricated
		// values so that self-referential types (a struct with an opt field
		// pointing transitively back to itself, as in the NNS governance
		// interface) terminate instead of recursing forever.
		return typeOfType(reflect.TypeOf(v), map[reflect.Type]*RecursiveType{})
	}
}

// typeOfType derives a candid Type from a Go reflect.Type. visited holds the
// recursive placeholders for struct types currently being expanded on the
// path from the root, so re-entering a type returns its placeholder instead
// of descending again.
func typeOfType(t reflect.Type, visited map[reflect.Type]*RecursiveType) (Type, error) {
	switch t.Kind() {
	case reflect.Bool:
		return new(BoolType), nil
	case reflect.Uint8:
		return Nat8Type(), nil
	case reflect.Uint16:
		return Nat16Type(), nil
	case reflect.Uint32:
		return Nat32Type(), nil
	case reflect.Uint, reflect.Uint64:
		return Nat64Type(), nil
	case reflect.Int8:
		return Int8Type(), nil
	case reflect.Int16:
		return Int16Type(), nil
	case reflect.Int32:
		return Int32Type(), nil
	case reflect.Int, reflect.Int64:
		return Int64Type(), nil
	case reflect.Float32:
		return Float32Type(), nil
	case reflect.Float64:
		return Float64Type(), nil
	case reflect.String:
		return new(TextType), nil
	case reflect.Slice, reflect.Array:
		if t == reflect.TypeOf([]byte(nil)) {
			// []byte is the natural Go representation of a candid blob (vec nat8).
			return NewVectorType(Nat8Type()), nil
		}
		elem, err := typeOfType(t.Elem(), visited)
		if err != nil {
			return nil, err
		}
		return NewVectorType(elem), nil
	case reflect.Pointer:
		// A pointer models a candid opt.
		elem, err := typeOfType(t.Elem(), visited)
		if err != nil {
			return nil, err
		}
		return NewOptionalType(elem), nil
	case reflect.Struct:
		// Special idl value-carrying structs (Nat, Int, Reserved, Empty, Null,
		// principal.Principal) are primitives, not records: defer to the
		// value-based TypeOf on a zero value so their canonical types are used.
		if isSpecialStruct(t) {
			return TypeOf(reflect.Zero(t).Interface())
		}
		if rec, ok := visited[t]; ok {
			rec.markUsed()
			return rec, nil // recursive back-reference
		}
		rec := NewRecursiveType(recursiveName(t))
		visited[t] = rec
		inner, err := structType(t, visited)
		if err != nil {
			return nil, err
		}
		delete(visited, t)
		// Only keep the recursive wrapper when the type actually refers back to
		// itself; otherwise return the plain type so non-recursive structs
		// encode identically to before.
		if !rec.Used() {
			return inner, nil
		}
		rec.setInner(inner)
		return rec, nil
	default:
		return nil, fmt.Errorf("unknown reflect type: %s", t)
	}
}

// isSpecialStruct reports whether t is one of the idl value-carrying structs
// that TypeOf treats as a primitive rather than a candid record.
func isSpecialStruct(t reflect.Type) bool {
	switch t {
	case reflect.TypeOf(Nat{}), reflect.TypeOf(Int{}),
		reflect.TypeOf(Reserved{}), reflect.TypeOf(Empty{}),
		reflect.TypeOf(Null{}), reflect.TypeOf(principal.Principal{}):
		return true
	}
	return false
}

// recursiveName is a stable, terminating name for a struct type used as the
// String() of its recursive placeholder.
func recursiveName(t reflect.Type) string {
	if t.Name() != "" {
		return "rec:" + t.PkgPath() + "." + t.Name()
	}
	return "rec:" + t.String()
}

func structType(t reflect.Type, visited map[reflect.Type]*RecursiveType) (Type, error) {
	variant := false
	tuple := false
	for i := range t.NumField() {
		f := t.Field(i)
		if !f.IsExported() {
			continue
		}
		tag := ParseTags(f)
		if tag.VariantType {
			variant = true
		}
		if tag.TupleType {
			tuple = true
		}
	}

	fields := make(map[string]Type)
	for i := range t.NumField() {
		f := t.Field(i)
		if !f.IsExported() {
			continue
		}
		tag := ParseTags(f)
		ft, err := typeOfType(f.Type, visited)
		if err != nil {
			return nil, err
		}
		if variant {
			// Variant arms are modelled as opt pointer fields; the arm type is
			// the pointed-to type.
			if o, ok := ft.(*OptionalType); ok {
				ft = o.Type
			} else {
				return nil, fmt.Errorf("variant field %q must be a pointer", tag.Name)
			}
		}
		fields[tag.Name] = ft
	}
	if variant {
		return NewVariantType(fields), nil
	}
	if tuple {
		return NewTupleType(fields), nil
	}
	return NewRecordType(fields), nil
}

type UnknownTypeError struct {
	Type Type
}

func (e UnknownTypeError) Error() string {
	return fmt.Sprintf("unknown idl type: %s", e.Type)
}

type UnknownValueTypeError struct {
	Value any
}

func (e UnknownValueTypeError) Error() string {
	return fmt.Sprintf("unknown idl value type: %s", reflect.TypeOf(e.Value))
}
