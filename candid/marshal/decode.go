package marshal

import (
	"errors"
	"fmt"
	"github.com/aviate-labs/agent-go/principal"
	"reflect"

	"github.com/aviate-labs/agent-go/candid/idl"
)

func Unmarshal(data []byte, values []any) error {
	ts, vs, err := idl.Decode(data)
	if err != nil {
		return err
	}
	if len(ts) != len(vs) {
		return fmt.Errorf("unequal data types and value lengths: %d %d", len(ts), len(vs))
	}

	if len(vs) != len(values) {
		return fmt.Errorf("unequal value lengths: %d %d", len(vs), len(values))
	}

	for i, v := range values {
		if err := unmarshal(ts[i], vs[i], v); err != nil {
			return err
		}
	}

	return nil
}

func recordMapToStruct(r *idl.RecordType, m map[string]any, value any) error {
	v := reflect.ValueOf(value).Elem()
	if !v.CanAddr() {
		return fmt.Errorf("can not address struct value")
	}

	fieldNameToIndex := make(map[string]int)
	for i := 0; i < v.NumField(); i++ {
		field := v.Type().Field(i)
		tag := idl.ParseTags(field)
		fieldNameToIndex[idl.HashString(tag.Name)] = i
	}

	for _, f := range r.Fields {
		value := m[f.Name]
		i := fieldNameToIndex[f.Name]
		if reflect.TypeOf(value) == v.Field(i).Type() {
			v.Field(i).Set(reflect.ValueOf(value))
			continue
		}
		fieldV := v.Field(i).Interface()
		if err := unmarshal(f.Type, value, &fieldV); err != nil {
			return err
		}
		v.Field(i).Set(reflect.ValueOf(fieldV))
	}
	return nil
}

func unmarshal(typ idl.Type, dv any, value any) error {
	v := reflect.ValueOf(value)
	if v.Kind() != reflect.Ptr || v.IsNil() {
		return errors.New("value is invalid")
	}

	if v.Type().Kind() == reflect.Ptr {
		v = v.Elem()
	}

	switch t := typ.(type) {
	case *idl.BoolType:
		switch v.Kind() {
		case reflect.Bool: // OK
		default:
			return NewErrInvalidTypeMatch(v, t)
		}
	case *idl.NatType:
		switch v.Kind() {
		case reflect.Uint8:
			if t.Base() != 1 {
				return fmt.Errorf("invalid base: %d, expected 1", t.Base())
			}
		case reflect.Uint16:
			if t.Base() != 2 {
				return fmt.Errorf("invalid base: %d, expected 2", t.Base())
			}
		case reflect.Uint32:
			if t.Base() != 4 {
				return fmt.Errorf("invalid base: %d, expected 4", t.Base())
			}
		case reflect.Uint64:
			if t.Base() != 8 {
				return fmt.Errorf("invalid base: %d, expected 8", t.Base())
			}
		case reflect.Struct:
			switch value.(type) {
			case *idl.Nat:
			default:
				return NewErrInvalidTypeMatch(v, t)
			}
		default:
			return NewErrInvalidTypeMatch(v, t)
		}
	case *idl.NullType:
		switch value.(type) {
		case *idl.Null:
		default:
			return NewErrInvalidTypeMatch(v, t)
		}
		return nil // No need to assign a value.
	case *idl.IntType:
		switch v.Kind() {
		case reflect.Int8:
			if t.Base() != 1 {
				return fmt.Errorf("invalid base: %d, expected 1", t.Base())
			}
		case reflect.Int16:
			if t.Base() != 2 {
				return fmt.Errorf("invalid base: %d, expected 2", t.Base())
			}
		case reflect.Int32:
			if t.Base() != 2 {
				return fmt.Errorf("invalid base: %d, expected 4", t.Base())
			}
		case reflect.Int64:
			if t.Base() != 2 {
				return fmt.Errorf("invalid base: %d, expected 8", t.Base())
			}
		case reflect.Struct:
			switch value.(type) {
			case *idl.Int:
			default:
				return NewErrInvalidTypeMatch(v, t)
			}
		default:
			return NewErrInvalidTypeMatch(v, t)
		}
	case *idl.FloatType:
		switch v.Kind() {
		case reflect.Float32:
			if t.Base() != 4 {
				return fmt.Errorf("invalid base: %d, expected 4", t.Base())
			}
		case reflect.Float64:
			if t.Base() != 8 {
				return fmt.Errorf("invalid base: %d, expected 8", t.Base())
			}
		default:
			return NewErrInvalidTypeMatch(v, t)
		}
	case *idl.TextType:
		switch v.Kind() {
		case reflect.String: // OK
		default:
			return NewErrInvalidTypeMatch(v, t)
		}
	case *idl.OptionalType:
		return typ.UnmarshalGo(dv, value)
	case *idl.VectorType:
		switch v.Kind() {
		case reflect.Slice:
		default:
			return NewErrInvalidTypeMatch(v, t)
		}
	case *idl.RecordType:
		switch v.Kind() {
		case reflect.Map:
		case reflect.Struct:
			return recordMapToStruct(t, dv.(map[string]any), value)
		default:
			return NewErrInvalidTypeMatch(v, t)
		}
	case *idl.VariantType:
		switch v.Kind() {
		case reflect.Struct:
			switch value.(type) {
			case *idl.Variant:
			default:
				return variantToStruct(t, dv.(*idl.Variant), value)
			}
		}
	case *idl.PrincipalType:
		switch v.Kind() {
		case reflect.Struct:
			switch value.(type) {
			case *principal.Principal:
			default:
				return NewErrInvalidTypeMatch(v, t)
			}
		}
	default:
		panic(fmt.Sprintf("%s, %v", typ, dv))
	}

	// Default behavior: there is no need to check/convert dv.
	v.Set(reflect.ValueOf(dv))
	return nil
}

func variantToStruct(r *idl.VariantType, variant *idl.Variant, value any) error {
	v := reflect.ValueOf(value).Elem()
	if !v.CanAddr() {
		return fmt.Errorf("can not address struct value")
	}

	fieldNameToIndex := make(map[string]int)
	for i := 0; i < v.NumField(); i++ {
		field := v.Type().Field(i)
		tag := idl.ParseTags(field)
		fieldNameToIndex[idl.HashString(tag.Name)] = i
	}

	i := fieldNameToIndex[r.Fields[0].Name]
	ptrValue := reflect.New(reflect.TypeOf(variant.Value))
	ptrValue.Elem().Set(reflect.ValueOf(variant.Value))
	v.Field(i).Set(ptrValue)
	return nil
}
