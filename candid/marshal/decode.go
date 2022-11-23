package marshal

import (
	"errors"
	"fmt"
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
			return fmt.Errorf("invalid type match: %s %s", v.Kind(), t)
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
			if t.Base() != 2 {
				return fmt.Errorf("invalid base: %d, expected 4", t.Base())
			}
		case reflect.Uint64:
			if t.Base() != 2 {
				return fmt.Errorf("invalid base: %d, expected 8", t.Base())
			}
		case reflect.Struct:
			switch v.Type().String() {
			case "idl.Nat":
			default:
				return fmt.Errorf("invalid type match: %s %s", v.Type(), t)
			}
		default:
			return fmt.Errorf("invalid type match: %s %s", v.Kind(), t)
		}
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
			switch v.Type().String() {
			case "idl.Int":
			default:
				return fmt.Errorf("invalid type match: %s %s", v.Type(), t)
			}
		default:
			return fmt.Errorf("invalid type match: %s %s", v.Kind(), t)
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
			return fmt.Errorf("invalid type match: %s %s", v.Kind(), t)
		}
	case *idl.TextType:
		switch v.Kind() {
		case reflect.String: // OK
		default:
			return fmt.Errorf("invalid type match: %s %s", v.Kind(), t)
		}
	case *idl.OptionalType:
		if dv == nil {
			return nil
		}
		dv = idl.Optional{
			V: dv,
			T: t.Type,
		}
	case *idl.VectorType:
		switch v.Kind() {
		case reflect.Slice:
		default:
			return fmt.Errorf("invalid type match: %s %s", v.Kind(), t)
		}
	case *idl.RecordType:
		switch v.Kind() {
		case reflect.Map:
		default:
			return fmt.Errorf("invalid type match: %s %s", v.Kind(), t)
		}
	case *idl.VariantType:
		switch v.Kind() {
		case reflect.Map:
		default:
			return fmt.Errorf("invalid type match: %s %s", v.Kind(), t)
		}
	case *idl.PrincipalType:
		switch v.Kind() {
		case reflect.Struct:
			switch v.Type().String() {
			case "principal.Principal":
			default:
				return fmt.Errorf("invalid type match: %s %s", v.Type(), t)
			}
		}
	default:
		panic(fmt.Sprintf("%s, %v", typ, dv))
	}

	// Default behavior: there is no need to check/convert dv.
	v.Set(reflect.ValueOf(dv))
	return nil
}
