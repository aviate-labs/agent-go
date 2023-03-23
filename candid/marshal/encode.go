package marshal

import (
	"fmt"
	"math/big"
	"reflect"

	"github.com/aviate-labs/agent-go/candid/idl"
	"github.com/aviate-labs/agent-go/principal"
	"github.com/aviate-labs/leb128"
)

func Marshal(args []any) ([]byte, error) {
	e := newEncodeState()
	types, err := types(args, e)
	if err != nil {
		return nil, err
	}
	data, err := values(args, e)
	if err != nil {
		return nil, err
	}
	return concat([]byte{'D', 'I', 'D', 'L'}, types, data), nil
}

func encode(v reflect.Value, tdt *idl.TypeDefinitionTable) ([]byte, []byte, error) {
	if v.Kind() == reflect.Interface {
		if v.IsNil() {
			return EncodeNull()
		}
		v = v.Elem()
	}
	switch v.Kind() {
	case reflect.Ptr:
		return encode(v.Elem(), tdt)
	case reflect.Bool:
		return EncodeBool(v.Bool())
	case reflect.Uint:
		return nil, nil, fmt.Errorf("use idl.Nat instead of uint")
	case reflect.Int:
		return nil, nil, fmt.Errorf("use idl.Int instead of int")
	case reflect.Uint8:
		return EncodeNat8(uint8(v.Uint()))
	case reflect.Uint16:
		return EncodeNat16(uint16(v.Uint()))
	case reflect.Uint32:
		return EncodeNat32(uint32(v.Uint()))
	case reflect.Uint64:
		return EncodeNat64(uint64(v.Uint()))
	case reflect.Int8:
		return EncodeInt8(int8(v.Uint()))
	case reflect.Int16:
		return EncodeInt16(int16(v.Uint()))
	case reflect.Int32:
		return EncodeInt32(int32(v.Uint()))
	case reflect.Int64:
		return EncodeInt64(int64(v.Uint()))
	case reflect.Float32:
		return EncodeFloat32(float32(v.Float()))
	case reflect.Float64:
		return EncodeFloat64(v.Float())
	case reflect.String:
		return EncodeText(v.String())
	case reflect.Slice:
		if a, ok := (v.Interface()).([]any); ok {
			return EncodeCons(a, tdt)
		}
		return nil, nil, fmt.Errorf("invalid array type: %s", v.Type())
	case reflect.Map:
		if m, ok := (v.Interface()).(map[string]any); ok {
			return EncodeCons(m, tdt)
		}
		return nil, nil, fmt.Errorf("invalid map type: %s", v.Type())
	case reflect.Struct:
		switch v.Type().String() {
		case "idl.Empty":
			return EncodeEmpty()
		case "idl.Int":
			bi := v.Interface().(idl.Int)
			return EncodeInt(bi)
		case "idl.Nat":
			bi := v.Interface().(idl.Nat)
			return EncodeNat(bi)
		case "idl.Null":
			return EncodeNull()
		case "idl.Reserved":
			return EncodeReserved()
		case "idl.Optional":
			v := v.Interface().(idl.Optional)
			return EncodeCons(v, tdt)
		case "idl.Variant":
			v := v.Interface().(idl.Variant)
			return EncodeCons(v, tdt)
		case "principal.Principal":
			p := v.Interface().(principal.Principal)
			return EncodePrincipal(p)
		}
		return nil, nil, fmt.Errorf("invalid struct type: %s", v.Type())
	default:
		return nil, nil, fmt.Errorf("invalid primary value: %s", v.Kind())
	}
}

func types(args []any, e *encodeState) ([]byte, error) {
	// T
	for _, a := range args {
		t, err := idl.TypeOf(a)
		if err != nil {
			return nil, err
		}
		if err := t.AddTypeDefinition(e.tdt); err != nil {
			return nil, err
		}
	}

	tdtl, err := leb128.EncodeUnsigned(big.NewInt(int64(len(e.tdt.Types))))
	if err != nil {
		return nil, err
	}
	var tdte []byte
	for _, t := range e.tdt.Types {
		tdte = append(tdte, t...)
	}
	return append(tdtl, tdte...), nil
}

func values(args []any, e *encodeState) ([]byte, error) {
	tsl, err := leb128.EncodeSigned(big.NewInt(int64(len(args))))
	if err != nil {
		return nil, err
	}
	var (
		ts []byte
		vs []byte
	)
	for _, v := range args {
		t, v, err := encode(reflect.ValueOf(v), e.tdt)
		if err != nil {
			return nil, err
		}
		ts = append(ts, t...)
		vs = append(vs, v...)
	}
	return concat(tsl, ts, vs), nil
}

type encodeState struct {
	tdt *idl.TypeDefinitionTable
}

func newEncodeState() *encodeState {
	return &encodeState{
		tdt: &idl.TypeDefinitionTable{
			Indexes: make(map[string]int),
		},
	}
}
