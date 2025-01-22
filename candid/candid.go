package candid

import (
	"fmt"
	"strings"

	"github.com/aviate-labs/agent-go/candid/did"
	"github.com/aviate-labs/agent-go/candid/idl"
	"github.com/aviate-labs/agent-go/candid/internal/candid"
	"github.com/aviate-labs/agent-go/candid/internal/candidvalue"
	"github.com/aviate-labs/agent-go/principal"
)

// DecodeValueString decodes the given value into a candid string.
func DecodeValueString(value []byte) (string, error) {
	types, values, err := idl.Decode(value)
	if err != nil {
		return "", err
	}
	if len(types) != 1 || len(values) != 1 {
		return "", fmt.Errorf("can not decode: %x", value)
	}
	s, err := valueToString(types[0], values[0])
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("(%s)", s), nil
}

// DecodeValuesString decodes the given values into a candid string.
func DecodeValuesString(types []idl.Type, values []any) (string, error) {
	var ss []string
	if len(types) != len(values) {
		return "", fmt.Errorf("unequal length")
	}
	for i := range types {
		s, err := valueToString(types[i], values[i])
		if err != nil {
			return "", err
		}
		ss = append(ss, s)
	}
	return fmt.Sprintf("(%s)", strings.Join(ss, "; ")), nil
}

// EncodeValueString encodes the given candid string into a byte slice.
func EncodeValueString(value string) ([]byte, error) {
	p, err := candidvalue.NewParser([]rune(value))
	if err != nil {
		return nil, err
	}
	n, err := p.ParseEOF(candidvalue.Values)
	if err != nil {
		return nil, err
	}
	types, args, err := did.ConvertValues(n)
	if err != nil {
		return nil, err
	}
	return idl.Encode(types, args)
}

// ParseDID parses the given raw .did files and returns the Program that is defined in it.
func ParseDID(raw []rune) (did.Description, error) {
	p, err := candid.NewParser(raw)
	if err != nil {
		return did.Description{}, err
	}
	n, err := p.ParseEOF(candid.Prog)
	if err != nil {
		return did.Description{}, err
	}
	return did.ConvertDescription(n), nil
}

func valueToString(typ idl.Type, value any) (string, error) {
	switch t := typ.(type) {
	case *idl.NullType:
		return "null", nil
	case *idl.BoolType:
		return fmt.Sprintf("%t", value), nil
	case *idl.NatType:
		return fmt.Sprintf("%v : %s", value, t.String()), nil
	case *idl.IntType:
		if t.Base() == 0 {
			return fmt.Sprintf("%v", value), nil
		}
		return fmt.Sprintf("%v : %s", value, t.String()), nil
	case *idl.FloatType:
		f, _ := value.(float64)
		return fmt.Sprintf("%.f : %s", f, t), nil
	case *idl.TextType:
		return fmt.Sprintf("%q", value), nil
	case *idl.ReservedType:
		return "reserved", nil
	case *idl.EmptyType:
		return "empty", nil
	case *idl.OptionalType:
		if value == nil {
			return "opt null", nil
		}
		s, err := valueToString(t.Type, value)
		if err != nil {
			return "", err
		}
		return fmt.Sprintf("opt %s", s), nil
	case *idl.VectorType:
		var ss []string
		for _, a := range value.([]any) {
			s, err := valueToString(t.Type, a)
			if err != nil {
				return "", err
			}
			ss = append(ss, s)
		}
		if len(ss) == 0 {
			return "vec {}", nil
		}
		return fmt.Sprintf("vec { %s }", strings.Join(ss, "; ")), nil
	case *idl.RecordType:
		var ss []string
		for _, f := range t.Fields {
			v := value.(map[string]any)
			s, err := valueToString(f.Type, v[f.Name])
			if err != nil {
				return "", nil
			}
			ss = append(ss, fmt.Sprintf("%s = %s", f.Name, s))
		}
		if len(ss) == 0 {
			return "record {}", nil
		}
		return fmt.Sprintf("record { %s }", strings.Join(ss, "; ")), nil
	case *idl.VariantType:
		f := t.Fields[0]
		v := value.(*idl.Variant).Value
		var s string
		switch t := f.Type.(type) {
		case *idl.NullType:
			s = f.Name
		default:
			sv, err := valueToString(t, v)
			if err != nil {
				return "", err
			}
			s = fmt.Sprintf("%s = %s", f.Name, sv)
		}
		return fmt.Sprintf("variant { %s }", s), nil
	case *idl.PrincipalType:
		p, _ := value.(principal.Principal)
		return fmt.Sprintf("principal %q", p), nil
	default:
		panic(fmt.Sprintf("%s, %v", typ, value))
	}
}
