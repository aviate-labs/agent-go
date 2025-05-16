package idl

import (
	"fmt"
	"reflect"

	"github.com/aviate-labs/agent-go/principal"
)

func EmptyOf(t Type) (any, error) {
	var v any = t
	if r := reflect.ValueOf(t); r.Kind() == reflect.Ptr {
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

		// Specific slices.
		switch t := reflect.TypeOf(v); t.Kind() {
		case reflect.Slice, reflect.Array:
			typ, err := TypeOf(reflect.New(t.Elem()).Elem().Interface())
			if err != nil {
				return nil, err
			}
			return NewVectorType(typ), nil
		case reflect.Struct:
			m, err := StructToMap(v)
			if err != nil {
				return nil, err
			}
			if isVariantType(v) {
				fields := make(map[string]Type)
				for k, v := range m {
					typ, err := TypeOf(v)
					if err != nil {
						return nil, err
					}
					switch t := typ.(type) {
					case *OptionalType:
						typ = t.Type
					default:
						return nil, UnknownValueTypeError{Value: v}
					}
					fields[k] = typ
				}
				return NewVariantType(fields), nil
			} else {
				fields := make(map[string]Type)
				for k, v := range m {
					typ, err := TypeOf(v)
					if err != nil {
						return nil, err
					}
					fields[k] = typ
				}
				if isTupleType(v) {
					return NewTupleType(fields), nil
				}
				return NewRecordType(fields), nil
			}
		case reflect.Ptr:
			indirect := reflect.Indirect(reflect.ValueOf(v))
			if !indirect.IsValid() {
				indirect = reflect.New(reflect.TypeOf(v).Elem())
				return TypeOf(indirect.Interface())
			}
			typ, err := TypeOf(indirect.Interface())
			if err != nil {
				return nil, err
			}
			return NewOptionalType(typ), nil
		default:
			return nil, UnknownValueTypeError{Value: v}
		}
	}
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
